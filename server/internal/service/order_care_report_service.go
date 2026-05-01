package service

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"github.com/google/uuid"
	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	_ "golang.org/x/image/webp"
)

//go:embed assets/order-care-report/base.jpg
var orderCareReportBaseImage []byte

//go:embed assets/order-care-report/SourceHanSansSC-Regular.otf
var orderCareReportFontFile []byte

var orderCareReportFontState struct {
	once sync.Once
	font *opentype.Font
	err  error
}

type OrderCareReportSectionInput struct {
	Checks []string `json:"checks"`
	Note   string   `json:"note"`
}

type CreateOrderCareReportInput struct {
	PetID        uint                        `json:"pet_id"`
	PortraitURL  string                      `json:"portrait_url"`
	Weight       string                      `json:"weight"`
	CareDate     string                      `json:"care_date"`
	NextCareDate string                      `json:"next_care_date"`
	CareContent  string                      `json:"care_content"`
	BodyShape    string                      `json:"body_shape"`
	Skin         OrderCareReportSectionInput `json:"skin"`
	Hair         OrderCareReportSectionInput `json:"hair"`
	Nails        OrderCareReportSectionInput `json:"nails"`
	EyesFace     OrderCareReportSectionInput `json:"eyes_face"`
	Ears         OrderCareReportSectionInput `json:"ears"`
	Oral         OrderCareReportSectionInput `json:"oral"`
	Anus         OrderCareReportSectionInput `json:"anus"`
}

type OrderCareReportResult struct {
	ImageURL string `json:"image_url"`
	ReportID uint   `json:"report_id"`
	BathDate string `json:"bath_date"`
}

type OrderCareReportService struct {
	orderRepo         *repository.OrderRepository
	petBathReportRepo *repository.PetBathReportRepository
}

func NewOrderCareReportService(orderRepo *repository.OrderRepository, petBathReportRepo *repository.PetBathReportRepository) *OrderCareReportService {
	return &OrderCareReportService{
		orderRepo:         orderRepo,
		petBathReportRepo: petBathReportRepo,
	}
}

func (s *OrderCareReportService) Create(shopID, orderID uint, input CreateOrderCareReportInput) (*OrderCareReportResult, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil || order.ShopID != shopID {
		return nil, errors.New("订单不存在")
	}

	if order.PayStatus != 1 {
		return nil, errors.New("仅已支付订单可生成报告")
	}

	selectedPet, ok := resolveSelectedOrderPet(order, input.PetID)
	if !ok {
		return nil, errors.New("所选猫咪不属于当前订单")
	}

	if input.PortraitURL == "" {
		return nil, errors.New("请先上传护理照片")
	}

	if input.NextCareDate == "" {
		return nil, errors.New("请填写建议下次护理日期")
	}

	bathDate, err := resolveOrderCareReportBathDate(order, input.CareDate)
	if err != nil {
		return nil, err
	}

	reportImage, err := renderOrderCareReport(order, selectedPet, input, bathDate)
	if err != nil {
		return nil, err
	}

	imageURL, err := saveRenderedOrderCareReport(reportImage)
	if err != nil {
		return nil, err
	}
	imagePath := orderCareReportUploadFilePath(imageURL)
	shouldCleanupImage := true
	defer func() {
		if shouldCleanupImage && imagePath != "" {
			_ = os.Remove(imagePath)
		}
	}()

	sortOrder, err := s.petBathReportRepo.GetNextSortOrder(shopID, input.PetID)
	if err != nil {
		return nil, err
	}

	report := &model.PetBathReport{
		ShopID:    shopID,
		PetID:     input.PetID,
		ImageURL:  imageURL,
		BathDate:  &bathDate,
		SortOrder: sortOrder,
	}
	if err := s.petBathReportRepo.Create(report); err != nil {
		return nil, err
	}
	shouldCleanupImage = false

	return &OrderCareReportResult{
		ImageURL: imageURL,
		ReportID: report.ID,
		BathDate: bathDate.Format("2006-01-02"),
	}, nil
}

func orderHasPet(order *model.Order, petID uint) bool {
	_, ok := resolveSelectedOrderPet(order, petID)
	return ok
}

func parseCareReportDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func resolveOrderCareReportBathDate(order *model.Order, careDate string) (time.Time, error) {
	if strings.TrimSpace(careDate) != "" {
		parsed, err := parseCareReportDate(careDate)
		if err != nil {
			return time.Time{}, errors.New("护理日期格式错误")
		}
		return parsed, nil
	}
	if order != nil && order.PayTime != nil {
		return *order.PayTime, nil
	}
	return time.Now(), nil
}

func renderOrderCareReport(order *model.Order, pet *model.Pet, input CreateOrderCareReportInput, bathDate time.Time) (image.Image, error) {
	baseImage, _, err := image.Decode(bytes.NewReader(orderCareReportBaseImage))
	if err != nil {
		return nil, fmt.Errorf("decode care report base image: %w", err)
	}

	dc := gg.NewContextForImage(baseImage)
	dc.SetRGB(0, 0, 0)

	if err := drawOrderCareReportPortrait(dc, input.PortraitURL); err != nil {
		return nil, err
	}

	drawData := buildOrderCareReportDrawData(pet, input, bathDate)
	if err := drawOrderCareReportPrimaryFields(dc, drawData); err != nil {
		return nil, err
	}
	drawOrderCareReportBodyShape(dc, drawData.BodyShape)
	drawOrderCareReportSections(dc, input)

	return dc.Image(), nil
}

func saveRenderedOrderCareReport(img image.Image) (string, error) {
	uploadPath := config.AppConfig.Upload.Path
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", err
	}

	filename := uuid.New().String() + ".jpg"
	dstPath := filepath.Join(uploadPath, filename)
	if err := gg.SaveJPG(dstPath, img, 92); err != nil {
		return "", err
	}
	return "/uploads/" + filename, nil
}

func orderCareReportUploadFilePath(imageURL string) string {
	uploadPath := config.AppConfig.Upload.Path
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	base := filepath.Base(imageURL)
	if base == "." || base == "" || base == string(filepath.Separator) {
		return ""
	}
	return filepath.Join(uploadPath, base)
}

type orderCareReportDrawData struct {
	PetName      string
	Breed        string
	Gender       string
	Age          string
	CareContent  string
	CareDate     string
	NextCareDate string
	Weight       string
	BodyShape    string
}

func buildOrderCareReportDrawData(pet *model.Pet, input CreateOrderCareReportInput, bathDate time.Time) orderCareReportDrawData {
	data := orderCareReportDrawData{
		CareContent:  compactOrderCareReportText(input.CareContent),
		CareDate:     bathDate.Format("2006-01-02"),
		NextCareDate: compactOrderCareReportText(input.NextCareDate),
		Weight:       normalizeOrderCareReportWeight(input.Weight),
		BodyShape:    normalizeOrderCareReportKey(input.BodyShape),
	}
	if pet == nil {
		return data
	}

	data.PetName = compactOrderCareReportText(pet.Name)
	data.Breed = compactOrderCareReportText(pet.Breed)
	data.Gender = orderCareReportGenderDisplay(pet.Gender)
	data.Age = orderCareReportAgeDisplay(pet.BirthDate, bathDate)

	return data
}

func drawOrderCareReportPortrait(dc *gg.Context, portraitURL string) error {
	uploadPath := config.AppConfig.Upload.Path
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	portraitPath := filepath.Join(uploadPath, filepath.Base(portraitURL))

	file, err := os.Open(portraitPath)
	if err != nil {
		return errors.New("护理照片不存在")
	}
	defer file.Close()

	portrait, _, err := image.Decode(file)
	if err != nil {
		return errors.New("护理照片读取失败")
	}

	bounds := portrait.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return errors.New("护理照片无效")
	}

	scale := math.Max(
		(orderCareReportPortraitFrame.Radius*2)/float64(bounds.Dx()),
		(orderCareReportPortraitFrame.Radius*2)/float64(bounds.Dy()),
	)

	dc.Push()
	dc.DrawCircle(orderCareReportPortraitFrame.CenterX, orderCareReportPortraitFrame.CenterY, orderCareReportPortraitFrame.Radius)
	dc.Clip()
	dc.Translate(orderCareReportPortraitFrame.CenterX, orderCareReportPortraitFrame.CenterY)
	dc.Scale(scale, scale)
	dc.DrawImageAnchored(portrait, 0, 0, 0.5, 0.5)
	dc.Pop()
	// gg 的 Push/Pop 不会恢复 Clip() 产生的 mask，必须手动清掉，
	// 否则后续正文和勾选都会继续被裁在头像圆框里。
	dc.ResetClip()

	return nil
}

func drawOrderCareReportPrimaryFields(dc *gg.Context, data orderCareReportDrawData) error {
	fields := []struct {
		Key     string
		Value   string
		Size    float64
		MinSize float64
	}{
		{Key: "pet_name", Value: data.PetName, Size: 38, MinSize: 24},
		{Key: "breed", Value: data.Breed, Size: 34, MinSize: 22},
		{Key: "gender", Value: data.Gender, Size: 30, MinSize: 22},
		{Key: "age", Value: data.Age, Size: 30, MinSize: 22},
		{Key: "care_content", Value: data.CareContent, Size: 32, MinSize: 20},
		{Key: "care_date", Value: data.CareDate, Size: 28, MinSize: 20},
		{Key: "next_care_date", Value: data.NextCareDate, Size: 28, MinSize: 20},
		{Key: "weight", Value: data.Weight, Size: 30, MinSize: 20},
	}

	for _, field := range fields {
		if err := drawCenteredOrderCareReportLineText(dc, orderCareReportPrimaryFieldBoxes[field.Key], field.Value, field.Size, field.MinSize, orderCareReportPrimaryFieldLimits[field.Key]); err != nil {
			return err
		}
	}

	return nil
}

func drawOrderCareReportBodyShape(dc *gg.Context, bodyShape string) {
	anchor, ok := orderCareReportBodyShapeAnchors[bodyShape]
	if !ok {
		return
	}
	drawOrderCareReportCheckmark(dc, anchor)
}

func drawOrderCareReportSections(dc *gg.Context, input CreateOrderCareReportInput) {
	sections := []struct {
		Name  string
		Input OrderCareReportSectionInput
	}{
		{Name: "skin", Input: input.Skin},
		{Name: "hair", Input: input.Hair},
		{Name: "nails", Input: input.Nails},
		{Name: "eyes_face", Input: input.EyesFace},
		{Name: "ears", Input: input.Ears},
		{Name: "oral", Input: input.Oral},
		{Name: "anus", Input: input.Anus},
	}

	for _, section := range sections {
		layout, ok := orderCareReportSectionLayouts[section.Name]
		if !ok {
			continue
		}
		for _, rawCheck := range section.Input.Checks {
			key, ok := canonicalOrderCareReportSectionCheck(section.Name, rawCheck)
			if !ok {
				continue
			}
			anchor, exists := layout.Checkboxes[key]
			if !exists {
				continue
			}
			drawOrderCareReportCheckmark(dc, anchor)
		}
		_ = drawLeftAlignedOrderCareReportLineText(dc, layout.NoteBox, section.Input.Note, 22, 18, layout.NoteLimit)
	}
}

func drawCenteredOrderCareReportLineText(dc *gg.Context, box orderCareReportLineBox, text string, size, minSize float64, runeLimit int) error {
	text = compactOrderCareReportText(text)
	if text == "" {
		return nil
	}

	widthLimit := box.Right - box.Left
	text = truncateOrderCareReportText(text, runeLimit)
	for currentSize := size; currentSize >= minSize; currentSize -= 2 {
		if err := setOrderCareReportFontFace(dc, currentSize); err != nil {
			return err
		}
		if width, _ := dc.MeasureString(text); width <= widthLimit {
			dc.DrawStringAnchored(text, (box.Left+box.Right)/2, box.Baseline, 0.5, 0)
			return nil
		}
	}

	if err := setOrderCareReportFontFace(dc, minSize); err != nil {
		return err
	}
	text = ellipsizeOrderCareReportText(dc, text, widthLimit)
	dc.DrawStringAnchored(text, (box.Left+box.Right)/2, box.Baseline, 0.5, 0)
	return nil
}

func drawLeftAlignedOrderCareReportLineText(dc *gg.Context, box orderCareReportLineBox, text string, size, minSize float64, runeLimit int) error {
	text = compactOrderCareReportText(text)
	if text == "" {
		return nil
	}

	widthLimit := box.Right - box.Left
	text = truncateOrderCareReportText(text, runeLimit)
	for currentSize := size; currentSize >= minSize; currentSize -= 2 {
		if err := setOrderCareReportFontFace(dc, currentSize); err != nil {
			return err
		}
		if width, _ := dc.MeasureString(text); width <= widthLimit {
			dc.DrawStringAnchored(text, box.Left, box.Baseline, 0, 0)
			return nil
		}
	}

	if err := setOrderCareReportFontFace(dc, minSize); err != nil {
		return err
	}
	text = ellipsizeOrderCareReportText(dc, text, widthLimit)
	dc.DrawStringAnchored(text, box.Left, box.Baseline, 0, 0)
	return nil
}

func drawOrderCareReportCheckmark(dc *gg.Context, point orderCareReportPoint) {
	dc.Push()
	dc.SetRGB(0, 0, 0)
	dc.SetLineWidth(4.2)
	dc.MoveTo(point.X-8, point.Y)
	dc.LineTo(point.X-1, point.Y+8)
	dc.LineTo(point.X+12, point.Y-8)
	dc.Stroke()
	dc.Pop()
}

func resolveSelectedOrderPet(order *model.Order, petID uint) (*model.Pet, bool) {
	if order == nil {
		return nil, false
	}
	if order.Pet != nil && order.Pet.ID == petID {
		return order.Pet, true
	}
	if order.PetID != nil && *order.PetID == petID {
		return nil, true
	}
	if order.Appointment != nil {
		for _, appointmentPet := range order.Appointment.Pets {
			if appointmentPet.PetID == petID {
				return appointmentPet.Pet, true
			}
		}
	}
	if order.FeedingPlan != nil {
		for _, feedingPet := range order.FeedingPlan.Pets {
			if feedingPet.PetID == petID {
				return feedingPet.Pet, true
			}
		}
	}
	for _, group := range order.PetGroups {
		if group.PetID != nil && *group.PetID == petID {
			return nil, true
		}
	}
	return nil, false
}

func orderCareReportGenderDisplay(gender int) string {
	switch gender {
	case 1:
		return "GG"
	case 2:
		return "MM"
	default:
		return ""
	}
}

func orderCareReportAgeDisplay(birthDate *time.Time, reference time.Time) string {
	if birthDate == nil {
		return ""
	}
	birth := birthDate.In(reference.Location())
	if birth.After(reference) {
		return ""
	}

	totalMonths := (reference.Year()-birth.Year())*12 + int(reference.Month()-birth.Month())
	if reference.Day() < birth.Day() {
		totalMonths--
	}
	if totalMonths < 0 {
		return ""
	}
	if totalMonths >= 12 {
		years := totalMonths / 12
		months := totalMonths % 12
		if months == 0 {
			return fmt.Sprintf("%d岁", years)
		}
		return fmt.Sprintf("%d岁%d月", years, months)
	}
	if totalMonths > 0 {
		return fmt.Sprintf("%d月", totalMonths)
	}

	days := int(reference.Sub(birth).Hours() / 24)
	if days <= 0 {
		return ""
	}
	return fmt.Sprintf("%d天", days)
}

func normalizeOrderCareReportWeight(value string) string {
	value = compactOrderCareReportText(value)
	replacer := strings.NewReplacer("kg", "", "KG", "", "Kg", "", "kG", "", "公斤", "")
	return strings.TrimSpace(replacer.Replace(value))
}

func compactOrderCareReportText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func truncateOrderCareReportText(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func ellipsizeOrderCareReportText(dc *gg.Context, value string, maxWidth float64) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}
	for len(runes) > 1 {
		candidate := string(runes) + "…"
		if width, _ := dc.MeasureString(candidate); width <= maxWidth {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}

func setOrderCareReportFontFace(dc *gg.Context, size float64) error {
	face, err := orderCareReportFontFace(size)
	if err != nil {
		return err
	}
	dc.SetFontFace(face)
	return nil
}

func orderCareReportFontFace(size float64) (font.Face, error) {
	orderCareReportFontState.once.Do(func() {
		orderCareReportFontState.font, orderCareReportFontState.err = opentype.Parse(orderCareReportFontFile)
	})
	if orderCareReportFontState.err != nil {
		return nil, orderCareReportFontState.err
	}
	face, err := opentype.NewFace(orderCareReportFontState.font, &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	})
	if err != nil {
		return nil, err
	}
	return face, nil
}

func canonicalOrderCareReportSectionCheck(sectionName, raw string) (string, bool) {
	aliases, ok := orderCareReportSectionCheckAliases[sectionName]
	if !ok {
		return "", false
	}
	key, exists := aliases[normalizeOrderCareReportKey(raw)]
	return key, exists
}

func normalizeOrderCareReportKey(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	return value
}

var orderCareReportSectionCheckAliases = map[string]map[string]string{
	"skin": {
		"normal":   "normal",
		"正常":       "normal",
		"dandruff": "dandruff",
		"皮屑":       "dandruff",
		"red":      "red",
		"redness":  "red",
		"发红":       "red",
		"greasy":   "greasy",
		"sticky":   "greasy",
		"黏腻":       "greasy",
		"粘腻":       "greasy",
		"scab":     "scab",
		"bump":     "scab",
		"lump":     "scab",
		"疙瘩":       "scab",
		"结痂":       "scab",
	},
	"hair": {
		"shedding":       "shedding",
		"brushed":        "shedding",
		"cleaned":        "shedding",
		"combed":         "shedding",
		"已梳理":            "shedding",
		"undercoat_many": "undercoat_many",
		"掉毛较多":           "undercoat_many",
		"dry":            "dry",
		"干燥":             "dry",
		"greasy":         "greasy",
		"sticky":         "greasy",
		"黏腻":             "greasy",
		"粘腻":             "greasy",
		"matting":        "matting",
		"knotted":        "matting",
		"matted":         "matting",
		"打结":             "matting",
	},
	"nails": {
		"trimmed":          "trimmed",
		"已修剪":              "trimmed",
		"dewclaw_abnormal": "dewclaw_abnormal",
		"abnormal_gap":     "dewclaw_abnormal",
		"toe_gap":          "dewclaw_abnormal",
		"趾间异常":             "dewclaw_abnormal",
		"pads_dry":         "pads_dry",
		"paw_pad_dry":      "pads_dry",
		"pad_dry":          "pads_dry",
		"足底干燥":             "pads_dry",
		"too_long":         "too_long",
		"overgrown":        "too_long",
		"过长":               "too_long",
		"wound":            "wound",
		"伤口":               "wound",
	},
	"eyes_face": {
		"cleaned":         "cleaned",
		"已清洁":             "cleaned",
		"已清洁理":            "cleaned",
		"tear_many":       "tear_many",
		"discharge":       "tear_many",
		"secretions":      "tear_many",
		"分泌物多":            "tear_many",
		"eye_red":         "eye_red",
		"red":             "eye_red",
		"redness":         "eye_red",
		"眼睛发红":            "eye_red",
		"eye_abnormal":    "eye_abnormal",
		"eyelid_abnormal": "eye_abnormal",
		"eyelid":          "eye_abnormal",
		"眼睑异常":            "eye_abnormal",
		"wound":           "wound",
		"伤口":              "wound",
	},
	"ears": {
		"cleaned":         "cleaned",
		"已清洁":             "cleaned",
		"touch_sensitive": "touch_sensitive",
		"sensitive":       "touch_sensitive",
		"讨厌被触摸":           "touch_sensitive",
		"inflamed":        "inflamed",
		"swollen":         "inflamed",
		"发红肿胀":            "inflamed",
		"earwax":          "earwax",
		"greasy":          "earwax",
		"ear_wax":         "earwax",
		"耳垢黏腻":            "earwax",
		"black_earwax":    "black_earwax",
		"black":           "black_earwax",
		"dark":            "black_earwax",
		"耳垢发黑":            "black_earwax",
	},
	"oral": {
		"normal":        "normal",
		"正常":            "normal",
		"tartar":        "tartar",
		"牙结石":           "tartar",
		"gum_red":       "gum_red",
		"gums_red":      "gum_red",
		"teeth_red":     "gum_red",
		"牙龈发红":          "gum_red",
		"gum_swollen":   "gum_swollen",
		"gums_swollen":  "gum_swollen",
		"teeth_swollen": "gum_swollen",
		"牙龈肿胀":          "gum_swollen",
	},
	"anus": {
		"normal":             "normal",
		"正常":                 "normal",
		"red":                "red",
		"redness":            "red",
		"发红":                 "red",
		"inflamed":           "inflamed",
		"swollen":            "inflamed",
		"鼓起肿胀":               "inflamed",
		"anal_gland_swollen": "anal_gland_swollen",
		"bulge":              "anal_gland_swollen",
		"腺体肿胀":               "anal_gland_swollen",
	},
}
