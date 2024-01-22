package ogame

// Item Is an ogame item that can be activated
type Item struct {
	Ref            string
	Name           string
	Image          string
	ImageLarge     string `json:"imageLarge"`
	Title          string
	Rarity         string // common
	Amount         int64
	AmountFree     int64 `json:"amount_free"`
	AmountBought   int64 `json:"amount_bought"`
	CanBeActivated bool  `json:"canBeActivated"`
	//Category                []string
	//Currency                string // dm
	Costs int64
	//IsReduced               bool
	//buyable                 bool
	//canBeBoughtAndActivated bool
	//isAnUpgrade             bool
	//isCharacterClassItem    bool
	//hasEnoughCurrency       bool
	//Cooldown                bool
	//extendable              bool
	//MoonOnlyItem            bool
	//duration                any
	//DurationExtension       any
	//TotalTime               any
	//timeLeft                any
	//status                  any
	//firstStatus             string // effecting
	//ToolTip                 string
	//buyTitle                string
	//activationTitle         string
}

// ActiveItem ...
type ActiveItem struct {
	ID            int64
	Ref           string
	Name          string
	TimeRemaining int64
	TotalDuration int64
	ImgSmall      string
}
