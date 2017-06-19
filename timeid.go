package event

import (
	"encoding/binary"
	//"fmt"
	"encoding/hex"
)

// GenerateTimeID TODO
func GenerateTimeID(time uint64, id uint64) ID {
	firstPart := make([]byte, 8)
	secondPart := make([]byte, 8)
	
	binary.BigEndian.PutUint64(firstPart, time)
	binary.BigEndian.PutUint64(secondPart, id)
	var total [16]byte 
	copy(total[:], firstPart[:8])
	copy(total[8:], secondPart[:8])
	return ID(total)
}

func (id ID) ToString() string {
	byteID := id.Byte()
	return hex.EncodeToString(byteID)
	// part1 := binary.LittleEndian.Uint64(byteID[:8])
	// part2 := binary.LittleEndian.Uint64(byteID[8:])
	// return fmt.Sprintf("%d%d", part1, part2)
}

func IDFromString(str string) ID {
	id, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	var idArray [16]byte
    copy(idArray[:], id[:16])
	return ID(idArray)
}

func (id ID) IDPart() uint64 {
	byteID := id.Byte()
	idPart := binary.BigEndian.Uint64(byteID[8:])
	return idPart
}

func (id ID) TimePart() uint64 {
	byteID := id.Byte()
	idPart := binary.BigEndian.Uint64(byteID[:8])
	return idPart
}

func FromtString(string) ID {
	return ID{}
}