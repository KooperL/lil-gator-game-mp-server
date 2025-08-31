package main

import (
	"encoding/json"
	"errors"
	"log"
)

type sessionKey = string
type displayName = string

type PlayerPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type PlayerRotation struct {
	Fx float64 `json:"fx"`
	Fy float64 `json:"fy"`
	Fz float64 `json:"fz"`
}

type PlayerAnimationState struct {
	Grounded       bool   `json:"grounded"`
	Climbing       bool   `json:"climbing"`
	Swimming       bool   `json:"swimming"`
	Gliding        bool   `json:"gliding"`
	Sledding       bool   `json:"sledding"`
	EquippedState  int    `json:"equippedState"`
	AttackTrigger  bool   `json:"attackTrigger"`
	RagdollTrigger bool   `json:"ragdollTrigger"`
	AnimationHash  string `json:"animationHash"`
}

type PlayerItems struct {
	HatItemID       string `json:"hatItemID"`
	LeftHandItemID  string `json:"leftHandItemID"`
	RightHandItemID string `json:"rightHandItemID"`
}

type PlayerVelocity struct {
	Speed         float64 `json:"speed"`
	VerticalSpeed float64 `json:"verticalSpeed"`
	Angle         float64 `json:"angle"`
}

type PlayerData struct {
	DisplayName displayName `json:"displayName"`
	WorldState  int         `json:"worldState"`
	PlayerPosition
	PlayerRotation
	PlayerAnimationState
	PlayerItems
	PlayerVelocity
}

type PlayerDataHub struct {
	PlayerStates  []PlayerData `json:"playerStates"`
	ServerVersion string       `json:"serverVersion"`
}

type PlayerClientPool = map[*Client]PlayerData

func (m PlayerData) fromJSONBytes(msg []byte) (PlayerData, error) {
	if err := json.Unmarshal(msg, &m); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
		return PlayerData{}, errors.New("Could not parse message!")
	}

	return m, nil
}

func (m PlayerDataHub) toJSONBytes() ([]byte, error) {
	return json.Marshal(m)
}
