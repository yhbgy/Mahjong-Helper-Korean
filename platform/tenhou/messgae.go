package tenhou

type message struct {
	Tag string `json:"tag" xml:"-"`

	//Name string `json:"name"` // id
	//Sex  string `json:"sx"`

	UserName string `json:"uname" xml:"-"`
	//RatingScale string `json:"ratingscale"`

	//N string `json:"n"`
	//J string `json:"j"`
	//G string `json:"g"`

	// 설명
	Seed   string `json:"seed" xml:"seed,attr"` // 설명
	Ten    string `json:"ten" xml:"ten,attr"`   // 설명
	Dealer string `json:"oya" xml:"oya,attr"`   // 설명
	Hai    string `json:"hai" xml:"hai,attr"`   // 설명
	Hai0   string `json:"-" xml:"hai0,attr"`
	Hai1   string `json:"-" xml:"hai1,attr"`
	Hai2   string `json:"-" xml:"hai2,attr"`
	Hai3   string `json:"-" xml:"hai3,attr"`

	// 설명

	// 설명
	Who  string `json:"who" xml:"who,attr"` // 설명
	Meld string `json:"m" xml:"m,attr"`     // 설명

	// 설명
	// 설명

	// 설명
	// 설명
	Step string `json:"step" xml:"step,attr"` // 1

	// 설명
	// 설명
	// 설명
	// `json:"step"` // 2

	// 설명
	T string `json:"t"` // 설명

	// 설명
	// ba, hai, m, machi, ten, yaku, doraHai, who, fromWho, sc
	//Ba string `json:"ba"` // 0,0
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명
	// 설명

	// 설명

	// 설명
	// type, lobby, gpid
	//Type  string `json:"type"`
	//Lobby string `json:"lobby"`
	//GPID  string `json:"gpid"`

	// 설명
	// `json:"seed"`
	// `json:"ten"`
	// `json:"oya"`
	// `json:"hai"`
	// 설명
	//Meld2    string `json:"m2"`
	//Meld3    string `json:"m3"`
	// 설명
	//Kawa1 string `json:"kawa1"`
	//Kawa2 string `json:"kawa2"`
	//Kawa3 string `json:"kawa3"`
}
