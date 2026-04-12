package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"gorm.io/gorm"
)

func roomGroupLabel(index int) string {
	if index < 1 {
		index = 1
	}
	return fmt.Sprintf("房间%d", index)
}

func sortBoardingRooms(rooms []model.BoardingOrderRoom) {
	sort.Slice(rooms, func(i, j int) bool {
		if rooms[i].RoomIndex == rooms[j].RoomIndex {
			return rooms[i].ID < rooms[j].ID
		}
		return rooms[i].RoomIndex < rooms[j].RoomIndex
	})
}

func activeBoardingRoomStatus(status string) bool {
	return status == model.BoardingOrderStatusPendingCheckin || status == model.BoardingOrderStatusCheckedIn
}

func legacyBoardingRoom(order *model.BoardingOrder) model.BoardingOrderRoom {
	room := model.BoardingOrderRoom{
		BoardingOrderID:        order.ID,
		CabinetID:              order.CabinetID,
		RoomIndex:              1,
		CheckInAt:              order.CheckInAt,
		CheckOutAt:             order.CheckOutAt,
		ActualCheckOutAt:       order.ActualCheckOutAt,
		Nights:                 order.Nights,
		BaseAmount:             order.BaseAmount,
		HolidaySurchargeAmount: order.HolidaySurchargeAmount,
		DiscountAmount:         order.DiscountAmount,
		ManualDiscountAmount:   order.ManualDiscountAmount,
		PayAmount:              order.PayAmount,
		Status:                 order.Status,
		PolicySnapshotJSON:     order.PolicySnapshotJSON,
		PriceSnapshotJSON:      order.PriceSnapshotJSON,
		Cabinet:                order.Cabinet,
		Pets:                   order.Pets,
	}
	return room
}

func displayBoardingRooms(order *model.BoardingOrder) []model.BoardingOrderRoom {
	if order == nil {
		return nil
	}
	if len(order.Rooms) == 0 {
		return []model.BoardingOrderRoom{legacyBoardingRoom(order)}
	}
	rooms := append([]model.BoardingOrderRoom(nil), order.Rooms...)
	sortBoardingRooms(rooms)
	return rooms
}

func parseBoardingPriceSnapshot(snapshot string) *BoardingPricePreview {
	if strings.TrimSpace(snapshot) == "" {
		return nil
	}
	var preview BoardingPricePreview
	if err := json.Unmarshal([]byte(snapshot), &preview); err != nil {
		return nil
	}
	return &preview
}

func cloneBoardingLines(lines []BoardingPriceLine) []BoardingPriceLine {
	if len(lines) == 0 {
		return nil
	}
	cloned := make([]BoardingPriceLine, len(lines))
	copy(cloned, lines)
	return cloned
}

func buildFallbackRoomPreview(room model.BoardingOrderRoom) BoardingRoomPreview {
	label := "寄养住宿"
	if room.Cabinet != nil && strings.TrimSpace(room.Cabinet.CabinetType) != "" {
		label = fmt.Sprintf("%s 寄养住宿", room.Cabinet.CabinetType)
	}
	petCount := len(room.Pets)
	if petCount < 1 {
		petCount = 1
	}
	lines := []BoardingPriceLine{
		{
			Type:      "base",
			Label:     label,
			Quantity:  maxInt(room.Nights, 1),
			UnitPrice: 0,
			Amount:    room.BaseAmount,
		},
	}
	if room.HolidaySurchargeAmount > 0 {
		lines = append(lines, BoardingPriceLine{
			Type:      "holiday_surcharge",
			Label:     "节假日加收",
			Quantity:  1,
			UnitPrice: room.HolidaySurchargeAmount,
			Amount:    room.HolidaySurchargeAmount,
		})
	}
	return BoardingRoomPreview{
		RoomIndex: maxInt(room.RoomIndex, 1),
		CabinetID: room.CabinetID,
		CabinetType: func() string {
			if room.Cabinet != nil {
				return room.Cabinet.CabinetType
			}
			return ""
		}(),
		PetCount:               petCount,
		CheckInAt:              room.CheckInAt,
		CheckOutAt:             room.CheckOutAt,
		Nights:                 room.Nights,
		BaseAmount:             room.BaseAmount,
		HolidaySurchargeAmount: room.HolidaySurchargeAmount,
		DiscountAmount:         room.DiscountAmount,
		ManualDiscountAmount:   room.ManualDiscountAmount,
		PayAmount:              roundMoney(maxBoardingFloat(room.PayAmount-room.ManualDiscountAmount, 0)),
		Lines:                  lines,
	}
}

func petIDsFromRoom(room model.BoardingOrderRoom) []uint {
	ids := make([]uint, 0, len(room.Pets))
	for _, pet := range room.Pets {
		if pet.PetID > 0 {
			ids = append(ids, pet.PetID)
		}
	}
	return ids
}

func buildAggregatePreviewFromRooms(customerID uint, rooms []model.BoardingOrderRoom) *BoardingPricePreview {
	if len(rooms) == 0 {
		return nil
	}
	orderedRooms := append([]model.BoardingOrderRoom(nil), rooms...)
	sortBoardingRooms(orderedRooms)

	roomPreviews := make([]BoardingRoomPreview, 0, len(orderedRooms))
	aggregate := &BoardingPricePreview{}
	var earliest string
	var latest string

	for _, room := range orderedRooms {
		rawPreview := parseBoardingPriceSnapshot(room.PriceSnapshotJSON)
		var roomPreview BoardingRoomPreview
		if rawPreview != nil {
			roomPreview = BoardingRoomPreview{
				RoomIndex: maxInt(room.RoomIndex, 1),
				CabinetID: room.CabinetID,
				CabinetType: func() string {
					if room.Cabinet != nil {
						return room.Cabinet.CabinetType
					}
					return ""
				}(),
				PetIDs:                 petIDsFromRoom(room),
				PetCount:               maxInt(rawPreview.PetCount, len(room.Pets)),
				CheckInAt:              room.CheckInAt,
				CheckOutAt:             room.CheckOutAt,
				Nights:                 rawPreview.Nights,
				RegularNights:          rawPreview.RegularNights,
				HolidayNights:          rawPreview.HolidayNights,
				BaseAmount:             rawPreview.BaseAmount,
				ExtraPetAmount:         rawPreview.ExtraPetAmount,
				HolidaySurchargeAmount: rawPreview.HolidaySurchargeAmount,
				DiscountAmount:         rawPreview.DiscountAmount,
				ManualDiscountAmount:   room.ManualDiscountAmount,
				PayAmount:              rawPreview.PayAmount,
				Lines:                  cloneBoardingLines(rawPreview.Lines),
			}
		} else {
			roomPreview = buildFallbackRoomPreview(room)
			roomPreview.PetIDs = petIDsFromRoom(room)
		}

		if roomPreview.PetCount < 1 {
			roomPreview.PetCount = maxInt(len(room.Pets), 1)
		}
		if roomPreview.CheckInAt == "" {
			roomPreview.CheckInAt = room.CheckInAt
		}
		if roomPreview.CheckOutAt == "" {
			roomPreview.CheckOutAt = room.CheckOutAt
		}

		if earliest == "" || roomPreview.CheckInAt < earliest {
			earliest = roomPreview.CheckInAt
		}
		if latest == "" || roomPreview.CheckOutAt > latest {
			latest = roomPreview.CheckOutAt
		}

		if room.Status == model.BoardingOrderStatusCancelled {
			roomPreview.ManualDiscountAmount = 0
			roomPreview.PayAmount = 0
			roomPreview.Lines = nil
			roomPreviews = append(roomPreviews, roomPreview)
			continue
		}

		aggregate.PetCount += roomPreview.PetCount
		aggregate.RegularNights += roomPreview.RegularNights
		aggregate.HolidayNights += roomPreview.HolidayNights
		aggregate.BaseAmount = roundMoney(aggregate.BaseAmount + roomPreview.BaseAmount)
		aggregate.ExtraPetAmount = roundMoney(aggregate.ExtraPetAmount + roomPreview.ExtraPetAmount)
		aggregate.HolidaySurchargeAmount = roundMoney(aggregate.HolidaySurchargeAmount + roomPreview.HolidaySurchargeAmount)
		aggregate.DiscountAmount = roundMoney(aggregate.DiscountAmount + roomPreview.DiscountAmount)
		aggregate.PayAmount = roundMoney(aggregate.PayAmount + roomPreview.PayAmount)

		for _, line := range roomPreview.Lines {
			aggregate.Lines = append(aggregate.Lines, BoardingPriceLine{
				Type:      line.Type,
				Label:     fmt.Sprintf("%s · %s", roomGroupLabel(roomPreview.RoomIndex), line.Label),
				Quantity:  line.Quantity,
				UnitPrice: line.UnitPrice,
				Amount:    line.Amount,
			})
		}
		roomPreviews = append(roomPreviews, roomPreview)
	}

	aggregate.CheckInAt = earliest
	aggregate.CheckOutAt = latest
	if earliest != "" && latest != "" {
		if start, err := time.Parse("2006-01-02", earliest); err == nil {
			if end, err := time.Parse("2006-01-02", latest); err == nil && end.After(start) {
				aggregate.Nights = int(end.Sub(start).Hours() / 24)
			}
		}
	}
	aggregate.Rooms = roomPreviews
	aggregate = applyMemberDiscountToBoardingPreview(customerID, aggregate)
	if len(aggregate.Rooms) > 0 {
		activeIndexes := make([]int, 0, len(aggregate.Rooms))
		totalRoomPay := 0.0
		for idx, roomPreview := range aggregate.Rooms {
			if orderedRooms[idx].Status == model.BoardingOrderStatusCancelled {
				continue
			}
			activeIndexes = append(activeIndexes, idx)
			totalRoomPay = roundMoney(totalRoomPay + roomPreview.PayAmount)
		}
		memberDiscountAmount := roundMoney(totalRoomPay - aggregate.PayAmount)
		if memberDiscountAmount > 0 && len(activeIndexes) > 0 {
			remaining := memberDiscountAmount
			for pos, idx := range activeIndexes {
				share := 0.0
				if pos == len(activeIndexes)-1 {
					share = remaining
				} else if totalRoomPay > 0 {
					share = roundMoney(memberDiscountAmount * aggregate.Rooms[idx].PayAmount / totalRoomPay)
					if share > remaining {
						share = remaining
					}
				}
				if share <= 0 {
					continue
				}
				remaining = roundMoney(remaining - share)
				aggregate.Rooms[idx].DiscountAmount = roundMoney(aggregate.Rooms[idx].DiscountAmount + share)
				aggregate.Rooms[idx].PayAmount = roundMoney(maxBoardingFloat(aggregate.Rooms[idx].PayAmount-share, 0))
				aggregate.Rooms[idx].Lines = append(aggregate.Rooms[idx].Lines, BoardingPriceLine{
					Type:      "member_discount",
					Label:     "会员折扣",
					Quantity:  1,
					UnitPrice: -share,
					Amount:    -share,
				})
			}
		}
	}
	for idx, roomPreview := range aggregate.Rooms {
		manualAmount := orderedRooms[idx].ManualDiscountAmount
		if manualAmount <= 0 {
			continue
		}
		manualAmount = roundMoney(minFloat(manualAmount, aggregate.PayAmount))
		aggregate.DiscountAmount = roundMoney(aggregate.DiscountAmount + manualAmount)
		aggregate.PayAmount = roundMoney(aggregate.PayAmount - manualAmount)
		aggregate.Rooms[idx].ManualDiscountAmount = manualAmount
		aggregate.Rooms[idx].PayAmount = roundMoney(maxBoardingFloat(roomPreview.PayAmount-manualAmount, 0))
		aggregate.Rooms[idx].Lines = append(aggregate.Rooms[idx].Lines, BoardingPriceLine{
			Type:      "manual_discount",
			Label:     "入住优惠",
			Quantity:  1,
			UnitPrice: -manualAmount,
			Amount:    -manualAmount,
		})
		aggregate.Lines = append(aggregate.Lines, BoardingPriceLine{
			Type:      "manual_discount",
			Label:     fmt.Sprintf("%s · 入住优惠", roomGroupLabel(aggregate.Rooms[idx].RoomIndex)),
			Quantity:  1,
			UnitPrice: -manualAmount,
			Amount:    -manualAmount,
		})
	}
	return aggregate
}

func summarizeBoardingOrderFromRooms(order *model.BoardingOrder, rooms []model.BoardingOrderRoom, preview *BoardingPricePreview) {
	if order == nil || len(rooms) == 0 {
		return
	}
	sortBoardingRooms(rooms)
	order.CabinetID = rooms[0].CabinetID
	order.CheckInAt = preview.CheckInAt
	order.CheckOutAt = preview.CheckOutAt
	order.Nights = preview.Nights
	order.BaseAmount = preview.BaseAmount
	order.HolidaySurchargeAmount = preview.HolidaySurchargeAmount
	order.DiscountAmount = preview.DiscountAmount
	order.ManualDiscountAmount = 0
	for _, room := range rooms {
		order.ManualDiscountAmount = roundMoney(order.ManualDiscountAmount + room.ManualDiscountAmount)
	}
	order.PayAmount = preview.PayAmount

	status := rooms[0].Status
	mixed := false
	allCheckedOut := true
	latestActual := ""
	for _, room := range rooms {
		if room.Status != status {
			mixed = true
		}
		if room.Status != model.BoardingOrderStatusCheckedOut {
			allCheckedOut = false
		}
		if room.ActualCheckOutAt > latestActual {
			latestActual = room.ActualCheckOutAt
		}
	}
	if mixed {
		order.Status = model.BoardingOrderStatusMixed
	} else {
		order.Status = status
	}
	if allCheckedOut {
		order.ActualCheckOutAt = latestActual
	} else {
		order.ActualCheckOutAt = ""
	}
}

func buildBoardingOrderItemsFromAggregate(orderID uint, preview *BoardingPricePreview) []model.OrderItem {
	if preview == nil {
		return nil
	}
	items := make([]model.OrderItem, 0, len(preview.Lines))
	for _, line := range preview.Lines {
		itemType := 4
		switch line.Type {
		case "holiday_surcharge":
			itemType = 5
		case "discount", "member_discount", "manual_discount":
			itemType = 6
		}
		if line.Amount == 0 {
			continue
		}
		items = append(items, model.OrderItem{
			OrderID:   orderID,
			ItemType:  itemType,
			ItemID:    0,
			Name:      line.Label,
			Quantity:  maxInt(line.Quantity, 1),
			UnitPrice: line.UnitPrice,
			Amount:    line.Amount,
		})
	}
	return items
}

func maxBoardingFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func applyAggregatePreviewToBoardingOrder(order *model.BoardingOrder, rooms []model.BoardingOrderRoom) (*BoardingPricePreview, error) {
	preview := buildAggregatePreviewFromRooms(order.CustomerID, rooms)
	if preview == nil {
		return nil, nil
	}
	summarizeBoardingOrderFromRooms(order, rooms, preview)
	snapshot, err := json.Marshal(preview)
	if err != nil {
		return nil, err
	}
	order.PriceSnapshotJSON = string(snapshot)
	return preview, nil
}

func syncBoardingPayOrder(tx *gorm.DB, order *model.BoardingOrder, preview *BoardingPricePreview, allowPaidCheckOut bool) error {
	if order == nil || preview == nil || order.OrderID == nil || *order.OrderID == 0 {
		return nil
	}
	var payOrder model.Order
	if err := tx.First(&payOrder, *order.OrderID).Error; err != nil {
		return err
	}
	if payOrder.PayStatus == 1 && !allowPaidCheckOut {
		return fmt.Errorf("已支付订单不可修改")
	}
	if payOrder.PayStatus == 1 && allowPaidCheckOut {
		return nil
	}
	if order.Status == model.BoardingOrderStatusCancelled {
		payOrder.TotalAmount = 0
		payOrder.ServiceTotal = 0
		payOrder.ProductTotal = 0
		payOrder.AddonTotal = 0
		payOrder.DiscountAmount = 0
		payOrder.ServiceDiscountAmount = 0
		payOrder.ProductDiscountAmount = 0
		payOrder.DiscountRate = 1
		payOrder.PayAmount = 0
		payOrder.Status = 2
		payOrder.PayStatus = 0
		if err := tx.Save(&payOrder).Error; err != nil {
			return err
		}
		return tx.Where("order_id = ?", payOrder.ID).Delete(&model.OrderItem{}).Error
	}
	payOrder.TotalAmount = roundMoney(preview.BaseAmount + preview.HolidaySurchargeAmount)
	payOrder.ServiceTotal = payOrder.TotalAmount
	payOrder.ProductTotal = 0
	payOrder.AddonTotal = 0
	payOrder.ServiceDiscountAmount = preview.DiscountAmount
	payOrder.ProductDiscountAmount = 0
	payOrder.DiscountAmount = preview.DiscountAmount
	payOrder.DiscountRate = calculateOrderDiscountRate(payOrder.TotalAmount, preview.PayAmount)
	payOrder.PayAmount = preview.PayAmount
	if err := tx.Save(&payOrder).Error; err != nil {
		return err
	}
	if err := tx.Where("order_id = ?", payOrder.ID).Delete(&model.OrderItem{}).Error; err != nil {
		return err
	}
	items := buildBoardingOrderItemsFromAggregate(payOrder.ID, preview)
	if len(items) == 0 {
		return nil
	}
	return tx.Create(&items).Error
}

func normalizeLoadedBoardingOrder(order *model.BoardingOrder) {
	if order == nil {
		return
	}
	rooms := displayBoardingRooms(order)
	order.Rooms = rooms
	if preview, err := applyAggregatePreviewToBoardingOrder(order, rooms); err == nil && preview != nil {
		if len(order.Pets) == 0 {
			for _, room := range rooms {
				order.Pets = append(order.Pets, room.Pets...)
			}
		}
	}
}

func (s *BoardingService) loadPersistedBoardingRooms(tx *gorm.DB, orderID uint) ([]model.BoardingOrderRoom, error) {
	var rooms []model.BoardingOrderRoom
	err := tx.Preload("Cabinet").
		Preload("Pets.Pet").
		Where("boarding_order_id = ?", orderID).
		Order("room_index ASC, id ASC").
		Find(&rooms).Error
	return rooms, err
}

func (s *BoardingService) refreshBoardingOrderAggregate(tx *gorm.DB, order *model.BoardingOrder) ([]model.BoardingOrderRoom, *BoardingPricePreview, error) {
	rooms, err := s.loadPersistedBoardingRooms(tx, order.ID)
	if err != nil {
		return nil, nil, err
	}
	preview, err := applyAggregatePreviewToBoardingOrder(order, rooms)
	if err != nil {
		return nil, nil, err
	}
	if err := tx.Save(order).Error; err != nil {
		return nil, nil, err
	}
	return rooms, preview, nil
}
