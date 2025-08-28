package main

import "encoding/json"
import "log"

type PlayerData struct {
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
	Z               float64 `json:"z"`
	Fx              float64 `json:"fx"`
	Fy              float64 `json:"fy"`
	Fz              float64 `json:"fz"`
	DisplayName     string  `json:"displayName"`
	SessionKey      string  `json:"sessionKey"`
	WorldState      int     `json:"worldState"`
	Speed           float64 `json:"speed"`
	VerticalSpeed   float64 `json:"verticalSpeed"`
	Angle           float64 `json:"angle"`
	Grounded        bool    `json:"grounded"`
	Climbing        bool    `json:"climbing"`
	Swimming        bool    `json:"swimming"`
	Gliding         bool    `json:"gliding"`
	Sledding        bool    `json:"sledding"`
  AttackTrigger   bool    `json:"attackTrigger"`
  RagdollTrigger  bool    `json:"ragdollTrigger"`
  EquippedState   int     `json:"equippedState"`
  HatItemID       string  `json:"hatItemID"`
  LeftHandItemID  string  `json:"leftHandItemID"`
  RightHandItemID string  `json:"rightHandItemID"`
}

func validMessage(msg []byte) bool {
  var player PlayerData
	if err := json.Unmarshal(msg, &player); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
    return false
	}

  return true
}
