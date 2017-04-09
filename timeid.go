package event

import (
	"encoding/binary"
)


func GenerateTimeID(time uint64, id uint64) ID {
	var firstPart, secondPart []byte
	
	binary.LittleEndian.PutUint64(firstPart, time)
	binary.LittleEndian.PutUint64(secondPart, id)
	var total [16]byte 
	copy(total[:], firstPart[:8])
	copy(total[8:], firstPart[:8])
	return ID(total)
}