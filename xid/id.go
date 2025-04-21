// Package xid provides utilities for ID generation, encoding, and zone-based ID management
package xid

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/speps/go-hashids/v2"
)

const (
	idStrLen = 18
	salt     = "fabrica2020"
	zoneBit  = 8
	// MaxZone is the maximum zone value (255) used in ID encoding
	MaxZone = (1 << zoneBit) - 1
)

var (
	h *hashids.HashID
)

func init() {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = idStrLen

	var err error
	if h, err = hashids.NewWithData(hd); err != nil {
		panic(errors.Wrapf(err, "hashID encode failed"))
	}
}

// CombineZoneID combines a zoneID with a zone value to create a combined ID
func CombineZoneID(zoneID int64, zone uint8) int64 {
	return (zoneID << zoneBit) | int64(zone)
}

// SplitID splits a combined ID into its zoneID and zone components
func SplitID(id int64) (zoneID int64, zone uint8) {
	zoneID = id >> zoneBit
	// Safely extract the zone bits without overflow risk
	zoneBits := id & 0xFF  // This is safe as 0xFF (255) is within int64 range
	zone = uint8(zoneBits) // This is safe as zoneBits is guaranteed to be 0-255

	return
}

// EncodeID encodes an ID into a string representation
// Returns the string ID or an error if encoding fails
func EncodeID(id int64) (string, error) {
	if id < 0 {
		return strconv.FormatInt(id, 10), nil
	}

	str, err := h.EncodeInt64([]int64{id})
	if err != nil {
		return "", errors.Wrapf(err, "HashID encode failed. id:%d", id)
	}

	return str, nil
}

// DecodeID decodes a string representation back into an ID
// Returns the decoded ID or an error if decoding fails
func DecodeID(str string) (int64, error) {
	if strings.IndexRune(str, '-') == 0 {
		return strconv.ParseInt(str, 10, 64)
	}

	ids, err := h.DecodeInt64WithError(str)
	if err != nil {
		return 0, errors.Wrapf(err, "HashID decode failed. str:%s", str)
	}

	if len(ids) == 0 {
		return 0, errors.Errorf("HashID decode failed. str:%s", str)
	}

	return ids[0], nil
}
