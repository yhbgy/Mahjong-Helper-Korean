// ==UserScript==
// @name         Mahjong Helper Korean - Majsoul WebSocket Hook
// @namespace    https://github.com/yhbgy/Mahjong-Helper-Korean
// @version      0.1.0
// @description  Forward Mahjong Soul WebSocket packets to the local Mahjong Helper server.
// @match        https://mahjongsoul.game.yo-star.com/kr/*
// @run-at       document-start
// @grant        unsafeWindow
// ==/UserScript==

(function () {
  "use strict";

  const pageWindow = typeof unsafeWindow !== "undefined" ? unsafeWindow : window;
  const RAW_ENDPOINT = "https://localhost:12121/majsoul-raw";
  const ACTION_ENDPOINT = "https://localhost:12121/majsoul";
  const LUA_PATCH_ENDPOINT = "https://localhost:12121/majsoul-lua-patches";
  const ACTION_BUNDLE_ENDPOINT = "https://localhost:12121/majsoul-action-bundle";
  const ACTION_NAMES = [
    "ActionNewRound",
    "ActionDealTile",
    "ActionDiscardTile",
    "ActionChiPengGang",
    "ActionAnGangAddGang",
    "ActionLiqi",
    "ActionHule",
    "ActionLiuJu",
    "ActionBabei",
    "ActionNoTile",
    "ActionMJStart",
  ];
  const ACTION_METHOD_NAMES = ["Play", "FastPlay", "Record", "FastRecord"];
  const LUA_BUNDLE_NAME = "2_ts1w92x_@036kahhn8x_9b0e5ee46a0df97a202a";
  const hookedActionObjects = new WeakSet();
  const NativeWebSocket = pageWindow.WebSocket;
  const NativeFetch = pageWindow.fetch;

  let luaPatchPromise = null;

  function shouldLogResourceUrl(url) {
    return /resources\/ab|clientbundlesettings|warehouseSettings|StreamingAssets|\.data(?:\.gz)?|\.unity3d|\.lua|\.bytes|\.json|bundle|majset/i.test(url);
  }

  function shouldProbeLuaBundleUrl(url) {
    return isActionBundleUrl(url) || /resources\/ab|\.data(?:\.gz)?|bundle|majset|\.bytes|\.bundle/i.test(url);
  }

  function urlOfRequest(input) {
    return typeof input === "string" ? input : input?.url;
  }

  function normalizedUrl(url) {
    try {
      return decodeURIComponent(url);
    } catch (_) {
      return url;
    }
  }

  function isActionBundleUrl(url) {
    return normalizedUrl(url).includes(LUA_BUNDLE_NAME);
  }

  function cacheBustUrl(url) {
    const separator = url.includes("?") ? "&" : "?";
    return `${url}${separator}mjh_nocache=${Date.now()}`;
  }

  function cacheBustFetchArgs(input, init) {
    const url = urlOfRequest(input);
    if (!url || !/\.data(?:\.gz)?(?:[?#].*)?$/i.test(url)) {
      return [input, init];
    }

    const nextInit = { ...(init || {}), cache: "reload" };
    if (typeof input === "string") {
      return [cacheBustUrl(input), nextInit];
    }

    return [new pageWindow.Request(cacheBustUrl(input.url), input), nextInit];
  }

  function base64ToBytes(value) {
    const binary = atob(value);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }

  function indexOfBytes(haystack, needle) {
    if (!needle.length || needle.length > haystack.length) {
      return -1;
    }

    const first = needle[0];
    const limit = haystack.length - needle.length;
    outer:
    for (let i = 0; i <= limit; i++) {
      if (haystack[i] !== first) continue;
      for (let j = 1; j < needle.length; j++) {
        if (haystack[i + j] !== needle[j]) continue outer;
      }
      return i;
    }
    return -1;
  }

  function readAscii(bytes, offset, length) {
    let out = "";
    for (let i = 0; i < length; i++) {
      out += String.fromCharCode(bytes[offset + i]);
    }
    return out;
  }

  function readU32LE(view, offset) {
    return view.getUint32(offset, true);
  }

  function parseUnityWebDataEntries(bytes) {
    const magic = "UnityWebData1.0\0";
    if (bytes.length < magic.length + 4 || readAscii(bytes, 0, magic.length) !== magic) {
      return null;
    }

    const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
    const headerEnd = readU32LE(view, magic.length);
    const entries = [];
    let pos = magic.length + 4;

    while (pos + 12 <= headerEnd && pos + 12 <= bytes.length) {
      const offset = readU32LE(view, pos);
      const size = readU32LE(view, pos + 4);
      const pathLength = readU32LE(view, pos + 8);
      pos += 12;
      if (pathLength <= 0 || pos + pathLength > headerEnd || offset + size > bytes.length) {
        break;
      }

      const path = readAscii(bytes, pos, pathLength);
      entries.push({ path, offset, size });
      pos += pathLength;
    }

    return entries;
  }

  function patchBytesInRange(bytes, patches, start, end) {
    let applied = 0;
    const missing = [];
    const range = bytes.subarray(start, end);

    for (const patch of patches) {
      const at = indexOfBytes(range, patch.original);
      if (at < 0) {
        missing.push(patch.file);
        continue;
      }
      bytes.set(patch.patched, start + at);
      applied++;
      console.log("[MJH] Lua patched:", patch.file);
    }

    return { applied, missing };
  }

  function getLuaPatches() {
    if (!luaPatchPromise) {
      luaPatchPromise = NativeFetch(LUA_PATCH_ENDPOINT)
        .then((response) => {
          if (!response.ok) {
            throw new Error(`patch endpoint returned ${response.status}`);
          }
          return response.json();
        })
        .then((patches) => patches.map((patch) => ({
          file: patch.file,
          original: base64ToBytes(patch.original_b64),
          patched: base64ToBytes(patch.patched_b64),
        })));
    }
    return luaPatchPromise;
  }

  async function patchLuaBundle(response, url) {
    const patches = await getLuaPatches();
    const bytes = new Uint8Array(await response.arrayBuffer());
    let applied = 0;
    const missing = [];
    const head = Array.from(bytes.slice(0, 16));
    const meta = {
      url,
      length: bytes.length,
      head,
      status: response.status,
      type: response.type,
      contentEncoding: response.headers?.get("Content-Encoding"),
      contentLength: response.headers?.get("Content-Length"),
    };

    if (bytes.length === 0) {
      console.log("[MJH] Lua patch empty response:", meta);
      return null;
    }

    const unityEntries = parseUnityWebDataEntries(bytes);
    if (unityEntries) {
      const interestingEntries = unityEntries.filter((entry) =>
        entry.path.includes(LUA_BUNDLE_NAME) ||
        /resources\/ab|bundle|majset|LuaByte|\.bytes/i.test(entry.path)
      );
      console.log("[MJH] UnityWebData entries:", {
        total: unityEntries.length,
        interesting: interestingEntries.slice(0, 40),
      });

      for (const entry of unityEntries) {
        const result = patchBytesInRange(bytes, patches, entry.offset, entry.offset + entry.size);
        if (result.applied > 0) {
          applied += result.applied;
          missing.push(...result.missing);
          console.log("[MJH] Lua patched in UnityWebData entry:", entry);
        }
      }
    } else {
      const result = patchBytesInRange(bytes, patches, 0, bytes.length);
      applied = result.applied;
      missing.push(...result.missing);
    }

    if (applied === 0) {
      console.log("[MJH] Lua patch probe miss:", meta);
      return null;
    } else {
      console.log("[MJH] Lua bundle patched:", { ...meta, applied, total: patches.length, missing });
    }

    return new pageWindow.Response(bytes, {
      status: response.status,
      statusText: response.statusText,
      headers: response.headers,
    });
  }

  if (NativeFetch) {
    pageWindow.fetch = function (input, init) {
      const url = urlOfRequest(input);
      if (url && shouldLogResourceUrl(url)) {
        console.log("[MJH] fetch:", url);
      }

      if (url && isActionBundleUrl(url)) {
        console.log("[MJH] redirect action bundle:", url);
        return NativeFetch(ACTION_BUNDLE_ENDPOINT, { cache: "no-store" }).then((response) => {
          console.log("[MJH] local action bundle response:", {
            status: response.status,
            ok: response.ok,
            type: response.type,
            contentLength: response.headers?.get("Content-Length"),
            contentType: response.headers?.get("Content-Type"),
          });
          return response;
        });
      }

      const fetchArgs = url && shouldProbeLuaBundleUrl(url)
        ? cacheBustFetchArgs(input, init)
        : arguments;
      const fetchResult = NativeFetch.apply(this, fetchArgs);
      if (url && shouldProbeLuaBundleUrl(url)) {
        return fetchResult
          .then((response) => patchLuaBundle(response.clone(), url).then((patched) => patched || response))
          .catch((err) => {
            console.warn("[MJH] Lua bundle patch failed:", err);
            return NativeFetch.apply(this, arguments);
          });
      }

      return fetchResult;
    };
  }

  if (pageWindow.Cache && pageWindow.Cache.prototype?.match) {
    const NativeCacheMatch = pageWindow.Cache.prototype.match;
    pageWindow.Cache.prototype.match = function (request, options) {
      const url = urlOfRequest(request);
      const matchResult = NativeCacheMatch.apply(this, arguments);
      if (!url || !shouldProbeLuaBundleUrl(url)) {
        return matchResult;
      }

      if (isActionBundleUrl(url)) {
        console.log("[MJH] redirect cached action bundle:", url);
        return NativeFetch(ACTION_BUNDLE_ENDPOINT, { cache: "no-store" }).then((response) => {
          console.log("[MJH] local cached action bundle response:", {
            status: response.status,
            ok: response.ok,
            type: response.type,
            contentLength: response.headers?.get("Content-Length"),
            contentType: response.headers?.get("Content-Type"),
          });
          return response;
        });
      }

      console.log("[MJH] cache match:", url);
      return matchResult.then((response) => {
        if (!response) return response;
        return patchLuaBundle(response.clone(), url).catch((err) => {
          console.warn("[MJH] cached Lua bundle patch failed:", err);
          return null;
        }).then((patched) => patched || response);
      });
    };
  }

  const NativeXHROpen = pageWindow.XMLHttpRequest.prototype.open;
  pageWindow.XMLHttpRequest.prototype.open = function (method, url) {
    if (url && shouldLogResourceUrl(url)) {
      console.log("[MJH] xhr:", method, url);
    }
    return NativeXHROpen.apply(this, arguments);
  };

  async function toArrayBuffer(data) {
    if (data instanceof ArrayBuffer) {
      return data;
    }

    if (ArrayBuffer.isView(data)) {
      return data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength);
    }

    if (data instanceof Blob) {
      return await data.arrayBuffer();
    }

    if (typeof data === "string") {
      return new TextEncoder().encode(data).buffer;
    }

    return null;
  }

  async function postRaw(direction, data) {
    const body = await toArrayBuffer(data);
    if (!body) return;

    fetch(`${RAW_ENDPOINT}?dir=${encodeURIComponent(direction)}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream",
      },
      body,
    }).catch((err) => {
      console.warn("[MJH] local post failed:", err);
    });
  }

  function sanitizeValue(value, depth = 0, seen = new WeakSet()) {
    if (value == null || typeof value === "string" || typeof value === "number" || typeof value === "boolean") {
      return value;
    }

    if (typeof value === "bigint") {
      return Number(value);
    }

    if (typeof value === "function" || typeof value === "symbol") {
      return undefined;
    }

    if (depth >= 6) {
      return String(value);
    }

    if (seen.has(value)) {
      return undefined;
    }
    seen.add(value);

    if (Array.isArray(value)) {
      return value.slice(0, 200).map((item) => sanitizeValue(item, depth + 1, seen));
    }

    if (ArrayBuffer.isView(value)) {
      return Array.from(value.slice ? value.slice(0, 64) : value).slice(0, 64);
    }

    if (value instanceof ArrayBuffer) {
      return { byteLength: value.byteLength };
    }

    const out = {};
    const keys = new Set();

    try {
      Object.keys(value).forEach((key) => keys.add(key));
    } catch (_) {
      // Some Unity/xLua proxy objects throw while enumerating; known fields below still get probed.
    }

    [
      "account_id",
      "seat_list",
      "ready_id_list",
      "is_game_start",
      "game_config",
      "chang",
      "ju",
      "ben",
      "tiles",
      "tiles0",
      "tiles1",
      "tiles2",
      "tiles3",
      "dora",
      "doras",
      "seat",
      "tile",
      "left_tile_count",
      "is_liqi",
      "is_wliqi",
      "moqie",
      "operation",
      "type",
      "froms",
      "hules",
      "scores",
      "md5",
    ].forEach((key) => keys.add(key));

    for (const key of keys) {
      try {
        const sanitized = sanitizeValue(value[key], depth + 1, seen);
        if (sanitized !== undefined) {
          out[key] = sanitized;
        }
      } catch (_) {
        // Ignore fields that cannot be read from Lua proxy values.
      }
    }

    return out;
  }

  function postAction(actionName, payload) {
    const body = sanitizeValue(payload);
    if (!body || typeof body !== "object") {
      console.warn("[MJH] action payload is not an object:", actionName, payload);
      return;
    }

    body._action = actionName;
    console.log("[MJH] action:", actionName, body);

    fetch(ACTION_ENDPOINT, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    }).catch((err) => {
      console.warn("[MJH] local action post failed:", err);
    });
  }

  function hookActionObject(actionName, actionObject) {
    if (!actionObject || hookedActionObjects.has(actionObject)) {
      return false;
    }

    let hooked = false;
    for (const methodName of ACTION_METHOD_NAMES) {
      const original = actionObject[methodName];
      if (typeof original !== "function" || original.__mjhHooked) {
        continue;
      }

      actionObject[methodName] = function (payload) {
        postAction(actionName, payload);
        return original.apply(this, arguments);
      };
      actionObject[methodName].__mjhHooked = true;
      hooked = true;
    }

    if (hooked) {
      hookedActionObjects.add(actionObject);
      console.log("[MJH] hooked action:", actionName);
    }
    return hooked;
  }

  function tryHookActions(root = pageWindow, path = "window", depth = 0, seen = new WeakSet()) {
    if (!root || typeof root !== "object" || seen.has(root) || depth > 3) {
      return 0;
    }
    seen.add(root);

    let count = 0;
    for (const actionName of ACTION_NAMES) {
      try {
        if (hookActionObject(actionName, root[actionName])) {
          count++;
        }
      } catch (_) {
        // Keep probing other roots.
      }
    }

    const likelyRoots = depth === 0
      ? ["Module", "unityInstance", "gameInstance", "app", "Game", "Laya", "CS", "xlua"]
      : [];
    for (const key of likelyRoots) {
      try {
        count += tryHookActions(root[key], `${path}.${key}`, depth + 1, seen);
      } catch (_) {
        // Ignore roots that cannot be inspected.
      }
    }

    return count;
  }

  function dumpMajsoulGlobals() {
    const matches = [];
    for (const key of Object.getOwnPropertyNames(pageWindow)) {
      if (/Action|Lua|XLua|Module|Unity|Game|Net|MJ/i.test(key)) {
        matches.push(key);
      }
    }
    console.log("[MJH] possible globals:", matches.sort());
    return matches;
  }

  function HookedWebSocket(url, protocols) {
    const ws = protocols === undefined
      ? new NativeWebSocket(url)
      : new NativeWebSocket(url, protocols);

    console.log("[MJH] WebSocket created:", url);

    const originalSend = ws.send;
    ws.send = function (data) {
      console.log("[MJH] send:", {
        type: Object.prototype.toString.call(data),
        length: data?.byteLength ?? data?.length ?? data?.size,
      });

      postRaw("send", data);
      return originalSend.call(this, data);
    };

    ws.addEventListener("message", (event) => {
      console.log("[MJH] receive:", {
        type: Object.prototype.toString.call(event.data),
        length: event.data?.byteLength ?? event.data?.length ?? event.data?.size,
      });

      postRaw("recv", event.data);
    });

    return ws;
  }

  HookedWebSocket.prototype = NativeWebSocket.prototype;
  HookedWebSocket.CONNECTING = NativeWebSocket.CONNECTING;
  HookedWebSocket.OPEN = NativeWebSocket.OPEN;
  HookedWebSocket.CLOSING = NativeWebSocket.CLOSING;
  HookedWebSocket.CLOSED = NativeWebSocket.CLOSED;

  pageWindow.WebSocket = HookedWebSocket;
  pageWindow.MJH = {
    postAction,
    tryHookActions,
    dumpMajsoulGlobals,
  };

  let probeCount = 0;
  const probeTimer = setInterval(() => {
    const hookedCount = tryHookActions();
    probeCount++;
    if (hookedCount > 0 || probeCount >= 120) {
      clearInterval(probeTimer);
      if (hookedCount === 0) {
        console.log("[MJH] no global action objects found; run MJH.dumpMajsoulGlobals() for clues");
      }
    }
  }, 1000);

  console.log("[MJH] WebSocket and action hook ready");
})();
