package service

const (
	orderCareReportWidth  = 1279
	orderCareReportHeight = 1810
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
	"pet_name":       {Left: 279, Right: 666, Baseline: 285},
	"breed":          {Left: 279, Right: 666, Baseline: 427},
	"gender":         {Left: 196, Right: 371, Baseline: 557},
	"age":            {Left: 490, Right: 668, Baseline: 557},
	"care_content":   {Left: 279, Right: 666, Baseline: 641},
	"care_date":      {Left: 279, Right: 666, Baseline: 775},
	"next_care_date": {Left: 905, Right: 1134, Baseline: 775},
	"weight":         {Left: 108, Right: 219, Baseline: 853},
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
	"thin":     {X: 408, Y: 833},
	"skinny":   {X: 573, Y: 833},
	"standard": {X: 737, Y: 833},
	"chubby":   {X: 901, Y: 833},
	"obese":    {X: 1066, Y: 833},
}

var orderCareReportSectionLayouts = map[string]orderCareReportSectionLayout{
	"skin": {
		Checkboxes: map[string]orderCareReportPoint{
			"normal":   {X: 408, Y: 928},
			"dandruff": {X: 573, Y: 928},
			"red":      {X: 737, Y: 928},
			"greasy":   {X: 901, Y: 928},
			"scab":     {X: 1066, Y: 928},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 981},
		NoteLimit: 30,
	},
	"hair": {
		Checkboxes: map[string]orderCareReportPoint{
			"shedding":       {X: 408, Y: 1027},
			"undercoat_many": {X: 573, Y: 1027},
			"dry":            {X: 737, Y: 1027},
			"greasy":         {X: 901, Y: 1027},
			"matting":        {X: 1066, Y: 1027},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1080},
		NoteLimit: 30,
	},
	"nails": {
		Checkboxes: map[string]orderCareReportPoint{
			"trimmed":          {X: 408, Y: 1125},
			"dewclaw_abnormal": {X: 573, Y: 1125},
			"pads_dry":         {X: 737, Y: 1125},
			"too_long":         {X: 901, Y: 1125},
			"wound":            {X: 1066, Y: 1125},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1178},
		NoteLimit: 30,
	},
	"eyes_face": {
		Checkboxes: map[string]orderCareReportPoint{
			"cleaned":      {X: 408, Y: 1222},
			"tear_many":    {X: 573, Y: 1222},
			"eye_red":      {X: 737, Y: 1222},
			"eye_abnormal": {X: 901, Y: 1222},
			"wound":        {X: 1066, Y: 1222},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1276},
		NoteLimit: 30,
	},
	"ears": {
		Checkboxes: map[string]orderCareReportPoint{
			"cleaned":         {X: 408, Y: 1321},
			"touch_sensitive": {X: 573, Y: 1321},
			"inflamed":        {X: 737, Y: 1321},
			"earwax":          {X: 901, Y: 1321},
			"black_earwax":    {X: 1066, Y: 1321},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1373},
		NoteLimit: 30,
	},
	"oral": {
		Checkboxes: map[string]orderCareReportPoint{
			"normal":      {X: 408, Y: 1420},
			"tartar":      {X: 737, Y: 1420},
			"gum_red":     {X: 901, Y: 1420},
			"gum_swollen": {X: 1066, Y: 1420},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1503},
		NoteLimit: 30,
	},
	"anus": {
		Checkboxes: map[string]orderCareReportPoint{
			"normal":             {X: 408, Y: 1605},
			"red":                {X: 737, Y: 1605},
			"inflamed":           {X: 901, Y: 1605},
			"anal_gland_swollen": {X: 1066, Y: 1605},
		},
		NoteBox:   orderCareReportLineBox{Left: 493, Right: 1166, Baseline: 1657},
		NoteLimit: 30,
	},
}
