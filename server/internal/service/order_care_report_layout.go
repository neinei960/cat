package service

const (
	orderCareReportWidth                = 1279
	orderCareReportHeight               = 1810
	orderCareReportCheckmarkStrokeWidth = 6.0
)

type orderCareReportLineBox struct {
	Left     float64
	Right    float64
	Baseline float64
}

type orderCareReportPoint struct {
	X float64
	Y float64
}

var (
	orderCareReportCheckmarkStartOffset = orderCareReportPoint{X: -10, Y: -1}
	orderCareReportCheckmarkKneeOffset  = orderCareReportPoint{X: -2, Y: 8}
	orderCareReportCheckmarkEndOffset   = orderCareReportPoint{X: 14, Y: -14}
)

type orderCareReportCircle struct {
	CenterX float64
	CenterY float64
	Radius  float64
}

type orderCareReportSectionLayout struct {
	Checkboxes map[string]orderCareReportPoint
	NoteBox    orderCareReportLineBox
	NoteLimit  int
}

var orderCareReportPortraitFrame = orderCareReportCircle{
	CenterX: 1000,
	CenterY: 355,
	Radius:  210,
}

var orderCareReportPrimaryFieldBoxes = map[string]orderCareReportLineBox{
	"pet_name":       {Left: 279, Right: 666, Baseline: 257},
	"breed":          {Left: 279, Right: 666, Baseline: 389},
	"gender":         {Left: 196, Right: 371, Baseline: 511},
	"age":            {Left: 490, Right: 668, Baseline: 511},
	"care_content":   {Left: 279, Right: 666, Baseline: 625},
	"care_date":      {Left: 279, Right: 666, Baseline: 730},
	"next_care_date": {Left: 905, Right: 1134, Baseline: 730},
	"weight":         {Left: 108, Right: 219, Baseline: 843},
}

var orderCareReportPrimaryFieldLimits = map[string]int{
	"pet_name":       14,
	"breed":          18,
	"gender":         4,
	"age":            8,
	"care_content":   18,
	"care_date":      12,
	"next_care_date": 12,
	"weight":         8,
}

var orderCareReportBodyShapeAnchors = map[string]orderCareReportPoint{
	"thin":     {X: 406, Y: 833},
	"skinny":   {X: 569, Y: 833},
	"standard": {X: 732, Y: 833},
	"chubby":   {X: 895, Y: 833},
	"obese":    {X: 1058, Y: 833},
}

var orderCareReportSectionLayouts = map[string]orderCareReportSectionLayout{
	"skin": {
		Checkboxes: map[string]orderCareReportPoint{
			"normal":   {X: 406, Y: 929},
			"dandruff": {X: 569, Y: 929},
			"red":      {X: 732, Y: 929},
			"greasy":   {X: 895, Y: 929},
			"scab":     {X: 1058, Y: 929},
			"wound":    {X: 406, Y: 977},
		},
		NoteBox:   orderCareReportLineBox{Left: 648, Right: 1166, Baseline: 975},
		NoteLimit: 80,
	},
	"hair": {
		Checkboxes: map[string]orderCareReportPoint{
			"shedding":       {X: 406, Y: 1025},
			"undercoat_many": {X: 569, Y: 1025},
			"dry":            {X: 732, Y: 1025},
			"greasy":         {X: 895, Y: 1025},
			"matting":        {X: 1058, Y: 1025},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1071},
		NoteLimit: 80,
	},
	"nails": {
		Checkboxes: map[string]orderCareReportPoint{
			"trimmed":          {X: 406, Y: 1121},
			"dewclaw_abnormal": {X: 569, Y: 1121},
			"pads_dry":         {X: 732, Y: 1121},
			"too_long":         {X: 895, Y: 1121},
			"wound":            {X: 1058, Y: 1121},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1168},
		NoteLimit: 80,
	},
	"eyes_face": {
		Checkboxes: map[string]orderCareReportPoint{
			"cleaned":      {X: 406, Y: 1217},
			"tear_many":    {X: 569, Y: 1217},
			"eye_red":      {X: 732, Y: 1217},
			"eye_abnormal": {X: 895, Y: 1217},
			"wound":        {X: 1058, Y: 1217},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1264},
		NoteLimit: 80,
	},
	"ears": {
		Checkboxes: map[string]orderCareReportPoint{
			"cleaned":         {X: 406, Y: 1313},
			"touch_sensitive": {X: 569, Y: 1313},
			"inflamed":        {X: 732, Y: 1313},
			"earwax":          {X: 895, Y: 1313},
			"black_earwax":    {X: 1058, Y: 1313},
			"wound":           {X: 406, Y: 1361},
		},
		NoteBox:   orderCareReportLineBox{Left: 648, Right: 1166, Baseline: 1359},
		NoteLimit: 80,
	},
	"oral": {
		Checkboxes: map[string]orderCareReportPoint{
			"normal":          {X: 406, Y: 1410},
			"touch_sensitive": {X: 569, Y: 1410},
			"tartar":          {X: 732, Y: 1410},
			"gum_red":         {X: 895, Y: 1410},
			"gum_swollen":     {X: 1058, Y: 1410},
			"oral_ulcer":      {X: 406, Y: 1458},
			"bad_breath":      {X: 569, Y: 1458},
			"dental_abnormal": {X: 732, Y: 1458},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1503},
		NoteLimit: 80,
	},
	"anus": {
		Checkboxes: map[string]orderCareReportPoint{
			"normal":   {X: 406, Y: 1553},
			"prolapse": {X: 569, Y: 1553},
			"red":      {X: 732, Y: 1553},
			"inflamed": {X: 895, Y: 1553},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1600},
		NoteLimit: 80,
	},
}
