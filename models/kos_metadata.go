package models

/*
Represents the full metadata for a Key Of Salvation (taken from Pinata).
*/
type KOSMetadata struct {
	Name         string      `json:"name"`
	Image        string      `json:"image"`
	AnimationUrl string      `json:"animation_url"`
	Attributes   []Attribute `json:"attributes,omitempty"`
}

/*
Represents a Key Of Salvation's metadata. A more simplified version compared to the `KOSMetadata` struct.
*/
type KOSSimplifiedMetadata struct {
	TokenID        int     `bson:"tokenID"`        // the token ID of the Key Of Salvation
	AnimationUrl   string  `bson:"animationUrl"`   // the animation URL of the Key Of Salvation
	HouseTrait     string  `bson:"houseTrait"`     // the house trait of the Key Of Salvation
	TypeTrait      string  `bson:"typeTrait"`      // the type trait of the Key Of Salvation
	LuckTrait      float64 `bson:"luckTrait"`      // the luck trait of the Key Of Salvation
	LuckBoostTrait float64 `bson:"luckBoostTrait"` // the luck boost trait of the Key Of Salvation
}

/*
The `Attribute` struct represents a single attribute for a Key Of Salvation.
*/
type Attribute struct {
	TraitType   string      `json:"trait_type,omitempty"`
	DisplayType string      `json:"display_type,omitempty"`
	Value       interface{} `json:"value,omitempty"`
}
