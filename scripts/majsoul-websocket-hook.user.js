// ==UserScript==
// @name         Mahjong Helper Korean - Majsoul WebSocket Hook
// @namespace    https://github.com/yhbgy/Mahjong-Helper-Korean
// @version      0.1.0
// @description  Forward Mahjong Soul WebSocket packets to the local Mahjong Helper server.
// @match        https://mahjongsoul.game.yo-star.com/kr/*
// @run-at       document-start
// @grant        none
// ==/UserScript==

(function () {
  "use strict";

  const LOCAL_ENDPOINT = "https://localhost:12121/majsoul-raw";
  const NativeWebSocket = window.WebSocket;

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

    fetch(`${LOCAL_ENDPOINT}?dir=${encodeURIComponent(direction)}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream",
      },
      body,
    }).catch((err) => {
      console.warn("[MJH] local post failed:", err);
    });
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

  window.WebSocket = HookedWebSocket;

  console.log("[MJH] WebSocket hook ready");
})();