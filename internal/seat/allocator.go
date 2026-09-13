package seat

import "sort"

func AllocateSeats(availableSeatNos []uint64, count int) ([]uint64, error) {
	if count != 1 && count != 2 {
		return nil, ErrInvalidSeatCount
	}

	seatNos := append([]uint64(nil), availableSeatNos...)
	sort.Slice(seatNos, func(i, j int) bool {
		return seatNos[i] < seatNos[j]
	})
	seatNos = uniqueSeatNos(seatNos)

	if len(seatNos) < count {
		return nil, ErrInsufficientSeats
	}

	if count == 1 {
		return []uint64{seatNos[0]}, nil
	}

	for i := 0; i < len(seatNos)-1; i++ {
		if seatNos[i]+1 == seatNos[i+1] {
			return []uint64{seatNos[i], seatNos[i+1]}, nil
		}
	}

	return []uint64{seatNos[0], seatNos[1]}, nil
}

func uniqueSeatNos(sortedSeatNos []uint64) []uint64 {
	if len(sortedSeatNos) == 0 {
		return sortedSeatNos
	}

	unique := sortedSeatNos[:1]
	for _, seatNo := range sortedSeatNos[1:] {
		if seatNo != unique[len(unique)-1] {
			unique = append(unique, seatNo)
		}
	}

	return unique
}
