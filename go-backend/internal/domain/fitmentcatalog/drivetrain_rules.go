package fitmentcatalog

// FreehubStandard identifies the cassette interface used by a rear hub.
type FreehubStandard string

const (
	FreehubStandardHG11         FreehubStandard = "HG-11"
	FreehubStandardHG           FreehubStandard = "HG"
	FreehubStandardHGL2         FreehubStandard = "HG-L2"
	FreehubStandardXDR          FreehubStandard = "XDR"
	FreehubStandardXD           FreehubStandard = "XD"
	FreehubStandardMicroSpline  FreehubStandard = "MICRO_SPLINE"
	FreehubStandardN3W          FreehubStandard = "N3W"
	FreehubStandardCampyClassic FreehubStandard = "CAMPY_CLASSIC"
)

const (
	spacerPositionInnerBase     = "inner_base"
	spacerPositionAdapterSleeve = "adapter_sleeve"
)

// SpacerPart describes one physical part of a spacer or adapter requirement.
// A compound requirement such as Shimano's 2.85 mm road cassette setup is
// represented by more than one part instead of hiding the parts in a string.
type SpacerPart struct {
	ThicknessMM float64 `json:"thickness_mm"`
	PartCode    string  `json:"part_code,omitempty"`
	Position    string  `json:"position,omitempty"`
	Description string  `json:"description"`
}

// SpacerRequirement describes the complete installation requirement for one
// cassette/freehub combination.
type SpacerRequirement struct {
	Required    bool         `json:"required"`
	ThicknessMM float64      `json:"thickness_mm"`
	PartCode    string       `json:"part_code,omitempty"`
	Position    string       `json:"position,omitempty"`
	Description string       `json:"description"`
	Parts       []SpacerPart `json:"parts,omitempty"`
}

// FreehubFitmentOption is one valid freehub choice for a cassette rule. Some
// cassettes have two valid choices (for example XD or XDR with a spacer), so
// the rule must not collapse those choices into a single display string.
type FreehubFitmentOption struct {
	Standard    FreehubStandard   `json:"standard"`
	DisplayName string            `json:"display_name"`
	Spacer      SpacerRequirement `json:"spacer"`
	ImageSrc    string            `json:"image_src"`
	Notes       string            `json:"notes,omitempty"`
}

// CassetteFitmentRule is the authoritative, read-only cassette compatibility
// record consumed by the calculation API and the SSR knowledge matrix.
type CassetteFitmentRule struct {
	RuleID             string                 `json:"rule_id"`
	Brand              string                 `json:"brand"`
	CassetteSpec       string                 `json:"cassette_spec"`
	DisplayName        string                 `json:"display_name"`
	HintGroupsets      string                 `json:"hint_groupsets"`
	Speed              int                    `json:"speed"`
	MinCogTeeth        int                    `json:"min_cog_teeth"`
	MaxCogTeeth        int                    `json:"max_cog_teeth"`
	RecommendedFreehub FreehubStandard        `json:"recommended_freehub"`
	FitmentOptions     []FreehubFitmentOption `json:"fitment_options"`
	ImageSrc           string                 `json:"image_src"`
	MechanicalNotes    string                 `json:"mechanical_notes"`
	RuleVersion        string                 `json:"rule_version"`
	SourceRefs         []string               `json:"-"` // Internal audit refs; never exposed to storefront users.
}

func noSpacer() SpacerRequirement {
	return SpacerRequirement{
		Description: "直接安装，无需垫圈或适配套件。",
	}
}

func spacer(thickness float64, description string) SpacerRequirement {
	return SpacerRequirement{
		Required:    true,
		ThicknessMM: thickness,
		Position:    spacerPositionInnerBase,
		Description: description,
		Parts: []SpacerPart{{
			ThicknessMM: thickness,
			Position:    spacerPositionInnerBase,
			Description: description,
		}},
	}
}

func compoundSpacer(description string, parts ...SpacerPart) SpacerRequirement {
	var total float64
	for _, part := range parts {
		total += part.ThicknessMM
	}
	return SpacerRequirement{
		Required:    true,
		ThicknessMM: total,
		Position:    spacerPositionInnerBase,
		Description: description,
		Parts:       parts,
	}
}

func adapterSleeve(description string) SpacerRequirement {
	return SpacerRequirement{
		Required:    true,
		ThicknessMM: 4.4,
		PartCode:    "AC21-N3W",
		Position:    spacerPositionAdapterSleeve,
		Description: description,
		Parts: []SpacerPart{{
			ThicknessMM: 4.4,
			PartCode:    "AC21-N3W",
			Position:    spacerPositionAdapterSleeve,
			Description: "安装 +4.4 mm 延长花键套筒，并使用 AC21-N3W 加长锁环。",
		}},
	}
}

func fitment(standard FreehubStandard, displayName string, spacerRequirement SpacerRequirement, imageSrc, notes string) FreehubFitmentOption {
	return FreehubFitmentOption{
		Standard:    standard,
		DisplayName: displayName,
		Spacer:      spacerRequirement,
		ImageSrc:    imageSrc,
		Notes:       notes,
	}
}

const (
	imageHG11Road     = "/public/wheelsetbuyersguide/choose freehub/shimano-8-9-10-11-speed-road-hyper-freehub.webp"
	imageHGMountain   = "/public/wheelsetbuyersguide/choose freehub/shimano-8-9-10-11-speed-mountain-hyper-freehub.webp"
	imageMicroSpline  = "/public/wheelsetbuyersguide/choose freehub/shimano-micro-spline-11-12-speed-mountain-freehub.webp"
	imageXD           = "/public/wheelsetbuyersguide/choose freehub/sram-xd-11-12-speed-mountain-freehub.webp"
	imageXDR          = "/public/wheelsetbuyersguide/choose freehub/sram-xdr-road-11-12-speed-freehub.webp"
	imageN3W          = "/public/wheelsetbuyersguide/choose freehub/Campagnolo-8-9-10-11-N3W-freehub.webp"
	imageCampyClassic = "/public/wheelsetbuyersguide/choose freehub/Campagnolo-8-9-10-11-spd-freehub.webp"
)

// DefaultDrivetrainRules returns the versioned rules shipped with the
// application. Keeping these rules in the domain package makes the first
// implementation deterministic and avoids introducing a database dependency
// into this read-only technical reference.
func DefaultDrivetrainRules() []CassetteFitmentRule {
	rules := []CassetteFitmentRule{
		{
			RuleID: "sram-12s-10-52t", Brand: "SRAM", CassetteSpec: "sram_12s_10_52t",
			DisplayName: "12速 10-52T / 10-50T", HintGroupsets: "GX / X01 / XX1 Eagle / Transmission",
			Speed: 12, MinCogTeeth: 10, MaxCogTeeth: 52, RecommendedFreehub: FreehubStandardXD,
			FitmentOptions: []FreehubFitmentOption{
				fitment(FreehubStandardXD, "SRAM XD（山地）", noSpacer(), imageXD, "XD 塔基直接安装，无垫圈。"),
				fitment(FreehubStandardXDR, "SRAM XDR（公路/Gravel）", spacer(1.85, "若使用 XDR 塔基，必须在塔基底座最内侧安装 1.85 mm 垫圈。"), imageXDR, "XDR 比 XD 长 1.85 mm。"),
			},
			ImageSrc: imageXD, MechanicalNotes: "10T 齿根直径小于传统 HG 外径，必须使用 XD/XDR 阶梯式收窄塔基。", RuleVersion: "v1.0",
		},
		{
			RuleID: "sram-12s-11-50t", Brand: "SRAM", CassetteSpec: "sram_12s_11_50t",
			DisplayName: "12速 11-50T", HintGroupsets: "NX / SX Eagle 及第三方山地 12 速",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 50, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{
				fitment(FreehubStandardHG11, "Shimano HG-11（公路长塔基）", spacer(1.85, "若装在公路 HG-11 长塔基上，需在塔基底座安装 1.85 mm 垫圈。"), imageHG11Road, "山地宽度飞轮在公路长塔基上需要补足轴向差。"),
				fitment(FreehubStandardHG, "Shimano HG（山地）", noSpacer(), imageHGMountain, "山地 HG 塔基可直接安装。"),
			},
			ImageSrc: imageHG11Road, MechanicalNotes: "11T 分体花键飞轮走传统 HG 接口。", RuleVersion: "v1.0",
		},
		{
			RuleID: "sram-12s-10-36t", Brand: "SRAM", CassetteSpec: "sram_12s_10_36t",
			DisplayName: "12速 10-28T / 10-33T / 10-36T", HintGroupsets: "RED / Force / Rival AXS 公路",
			Speed: 12, MinCogTeeth: 10, MaxCogTeeth: 36, RecommendedFreehub: FreehubStandardXDR,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardXDR, "SRAM XDR", noSpacer(), imageXDR, "原生 XDR 公路飞轮无需垫圈。")},
			ImageSrc:       imageXDR, MechanicalNotes: "原厂 10T 一体飞轮直接锁紧在 XDR 塔基。", RuleVersion: "v1.0",
		},
		{
			RuleID: "sram-11s-10-42t", Brand: "SRAM", CassetteSpec: "sram_11s_10_42t",
			DisplayName: "11速 10-42T", HintGroupsets: "SRAM 11 速山地 XD 飞轮",
			Speed: 11, MinCogTeeth: 10, MaxCogTeeth: 42, RecommendedFreehub: FreehubStandardXD,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardXD, "SRAM XD", noSpacer(), imageXD, "直接安装，无垫圈。")},
			ImageSrc:       imageXD, MechanicalNotes: "10T 起跳必须使用 XD 接口。", RuleVersion: "v1.0",
		},
		{
			RuleID: "sram-11s-11-36t", Brand: "SRAM", CassetteSpec: "sram_11s_11_36t",
			DisplayName: "11速 11-28T 至 11-36T", HintGroupsets: "SRAM 11 速公路 PG 飞轮",
			Speed: 11, MinCogTeeth: 11, MaxCogTeeth: 36, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11", noSpacer(), imageHG11Road, "直接安装，无垫圈。")},
			ImageSrc:       imageHG11Road, MechanicalNotes: "11T 起跳的传统花键飞轮兼容 HG-11。", RuleVersion: "v1.0",
		},
		{
			RuleID: "shimano-12s-road-11-34t", Brand: "Shimano", CassetteSpec: "shimano_12s_road_11_34t",
			DisplayName: "12速公路 11-30T / 11-34T", HintGroupsets: "Dura-Ace R9200 / Ultegra R8100 / 105 R7100",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 34, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{
				fitment(FreehubStandardHG11, "Shimano HG-11", noSpacer(), imageHG11Road, "可直接安装在标准 11 速公路 HG 塔基。"),
				fitment(FreehubStandardHGL2, "Shimano HG L2", noSpacer(), imageHG11Road, "HG L2 同样兼容 Shimano 12 速公路飞轮。"),
			},
			ImageSrc: imageHG11Road, MechanicalNotes: "Shimano 12 速公路飞轮以 11T 起跳，向下兼容 HG-11；严禁装 Micro Spline。", RuleVersion: "v1.0",
		},
		{
			RuleID: "shimano-12s-mtb-10-51t", Brand: "Shimano", CassetteSpec: "shimano_12s_mtb_10_51t",
			DisplayName: "12速山地 10-45T / 10-51T", HintGroupsets: "XTR M9100 / XT M8100 / SLX M7100 / Deore M6100",
			Speed: 12, MinCogTeeth: 10, MaxCogTeeth: 51, RecommendedFreehub: FreehubStandardMicroSpline,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardMicroSpline, "Shimano Micro Spline", noSpacer(), imageMicroSpline, "23 花键微型塔基，直接安装。")},
			ImageSrc:       imageMicroSpline, MechanicalNotes: "10T 齿根干涉 HG，必须使用 Micro Spline。", RuleVersion: "v1.0",
		},
		{
			RuleID: "shimano-11s-road-11-34t", Brand: "Shimano", CassetteSpec: "shimano_11s_road_11_34t",
			DisplayName: "11速公路 11-25T 至 11-34T", HintGroupsets: "Dura-Ace R9100 / Ultegra R8000 / 105 R7000",
			Speed: 11, MinCogTeeth: 11, MaxCogTeeth: 34, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11", noSpacer(), imageHG11Road, "直接安装，无垫圈。")},
			ImageSrc:       imageHG11Road, MechanicalNotes: "行业标准 11 速公路长塔基。", RuleVersion: "v1.0",
		},
		{
			RuleID: "shimano-11s-mtb-11-42t", Brand: "Shimano", CassetteSpec: "shimano_11s_mtb_11_42t",
			DisplayName: "11速山地 11-40T / 11-42T / 11-46T", HintGroupsets: "XT M8000 / SLX M7000",
			Speed: 11, MinCogTeeth: 11, MaxCogTeeth: 46, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11（公路长塔基）", spacer(1.85, "装在公路 HG-11 长塔基上需安装 1.85 mm 垫圈。"), imageHG11Road, "山地 11 速飞轮底座为 10 速宽度。")},
			ImageSrc:       imageHG11Road, MechanicalNotes: "山地飞轮背面悬臂设计导致公路长塔基需要 1.85 mm 垫圈。", RuleVersion: "v1.0",
		},
		{
			RuleID: "shimano-10s-road-11-30t", Brand: "Shimano", CassetteSpec: "shimano_10s_road_11_30t",
			DisplayName: "10速公路 11/12-28T / 11-30T", HintGroupsets: "CS-6700 / CS-5700",
			Speed: 10, MinCogTeeth: 11, MaxCogTeeth: 30, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11", compoundSpacer("需要 1.85 mm 塔基转换垫圈 + 飞轮自带 1.0 mm 垫圈，共 2.85 mm。", SpacerPart{ThicknessMM: 1.85, Position: spacerPositionInnerBase, Description: "塔基底座转换垫圈。"}, SpacerPart{ThicknessMM: 1.0, Position: spacerPositionInnerBase, Description: "飞轮背面原装 1.0 mm 垫圈。"}), imageHG11Road, "必须同时安装两片垫圈。")},
			ImageSrc:       imageHG11Road, MechanicalNotes: "塔基转换差 1.85 mm 加上飞轮背部凹槽 1.0 mm。", RuleVersion: "v1.0",
		},
		{
			RuleID: "shimano-10s-tiagra-11-30t", Brand: "Shimano", CassetteSpec: "shimano_10s_tiagra_11_30t",
			DisplayName: "Tiagra 10速 11/12-28T / 11-30T", HintGroupsets: "CS-4600 / CS-HG500-10",
			Speed: 10, MinCogTeeth: 11, MaxCogTeeth: 30, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11", spacer(1.85, "安装在 HG-11 长塔基上需 1.85 mm 垫圈。"), imageHG11Road, "平底飞轮无需额外 1.0 mm 垫圈。")},
			ImageSrc:       imageHG11Road, MechanicalNotes: "Tiagra/HG500 平底设计仅需 1.85 mm。", RuleVersion: "v1.0",
		},
		{
			RuleID: "campagnolo-13s-ekar", Brand: "Campagnolo", CassetteSpec: "campagnolo_13s_ekar_9_44t",
			DisplayName: "13速 9-42T / 10-44T", HintGroupsets: "Ekar Gravel/Road",
			Speed: 13, MinCogTeeth: 9, MaxCogTeeth: 44, RecommendedFreehub: FreehubStandardN3W,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardN3W, "Campagnolo N3W", noSpacer(), imageN3W, "原生 N3W 飞轮直接安装，严禁加装套筒。")},
			ImageSrc:       imageN3W, MechanicalNotes: "N3W 紧凑塔基原生容纳 9T/10T 最小齿片。", RuleVersion: "v1.0",
		},
		{
			RuleID: "campagnolo-11-12s-11-34t", Brand: "Campagnolo", CassetteSpec: "campagnolo_11_12s_11_34t",
			DisplayName: "11/12速 11-29T 至 11-34T", HintGroupsets: "Super Record / Record / Chorus",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 34, RecommendedFreehub: FreehubStandardN3W,
			FitmentOptions: []FreehubFitmentOption{
				fitment(FreehubStandardN3W, "Campagnolo N3W", adapterSleeve("传统 11/12 速飞轮装在 N3W 上，需使用 AC21-N3W 套筒和加长锁环。"), imageN3W, "N3W 比经典塔基短 4.4 mm。"),
				fitment(FreehubStandardCampyClassic, "Campagnolo Classic", noSpacer(), imageCampyClassic, "经典塔基直接安装。"),
			},
			ImageSrc: imageN3W, MechanicalNotes: "N3W 装传统 11/12 速飞轮必须补足 4.4 mm 轴向长度。", RuleVersion: "v1.0",
		},
		{
			RuleID: "campagnolo-classic-9-12s", Brand: "Campagnolo", CassetteSpec: "campagnolo_classic_9_12s_11t",
			DisplayName: "9-12速 11T 起跳", HintGroupsets: "Campagnolo Classic cassette",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 34, RecommendedFreehub: FreehubStandardCampyClassic,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardCampyClassic, "Campagnolo Classic", noSpacer(), imageCampyClassic, "直接安装，无套筒。")},
			ImageSrc:       imageCampyClassic, MechanicalNotes: "传统 Campagnolo 长塔基兼容经典飞轮。", RuleVersion: "v1.0",
		},
		{
			RuleID: "ltwoo-road-12s-11-34t", Brand: "L-TWOO", CassetteSpec: "ltwoo_12s_road_11_34t",
			DisplayName: "公路12速 11-32T / 11-34T", HintGroupsets: "eRX / RX / R9 12S",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 34, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11", noSpacer(), imageHG11Road, "直接安装，无需换塔基。")}, ImageSrc: imageHG11Road, MechanicalNotes: "国产公路 12 速采用 11T 起跳 HG 花键。", RuleVersion: "v1.0",
		},
		{
			RuleID: "ltwoo-mtb-12s-11-52t", Brand: "L-TWOO", CassetteSpec: "ltwoo_12s_mtb_11_52t",
			DisplayName: "山地12速 11-50T / 11-52T", HintGroupsets: "A12 12S",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 52, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11（公路轮组）", spacer(1.85, "若安装在公路长塔基上，需 1.85 mm 垫圈。"), imageHG11Road, "山地塔基宽度可直接安装。")}, ImageSrc: imageHG11Road, MechanicalNotes: "山地宽度飞轮装公路长塔基需补足轴向差。", RuleVersion: "v1.0",
		},
		{
			RuleID: "sensah-road-12s-11-34t", Brand: "SENSAH", CassetteSpec: "sensah_12s_road_11_34t",
			DisplayName: "公路12速 11-32T / 11-34T", HintGroupsets: "Empire Pro",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 34, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11", noSpacer(), imageHG11Road, "直接锁紧，无垫圈。")}, ImageSrc: imageHG11Road, MechanicalNotes: "按 Shimano HG 接口设计。", RuleVersion: "v1.0",
		},
		{
			RuleID: "sensah-gravel-11-12s-11-50t", Brand: "SENSAH", CassetteSpec: "sensah_gravel_11_12s_11_50t",
			DisplayName: "Gravel 11/12速 11-42T / 11-50T", HintGroupsets: "SRX Pro",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 50, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11（公路轮组）", spacer(1.85, "若安装在公路长塔基上，需 1.85 mm 垫圈。"), imageHG11Road, "山地宽度飞轮需要垫圈。")}, ImageSrc: imageHG11Road, MechanicalNotes: "标准山地 HG 飞轮接口。", RuleVersion: "v1.0",
		},
		{
			RuleID: "microshift-sword-10s-11-48t", Brand: "microSHIFT", CassetteSpec: "microshift_sword_10s_11_48t",
			DisplayName: "Sword Gravel 10速 11-48T", HintGroupsets: "Sword",
			Speed: 10, MinCogTeeth: 11, MaxCogTeeth: 48, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11（公路轮组）", spacer(1.85, "若安装在公路长塔基上，需 1.85 mm 垫圈。"), imageHG11Road, "无需 XD 或 Micro Spline。")}, ImageSrc: imageHG11Road, MechanicalNotes: "宽齿比但仍采用传统 HG 接口。", RuleVersion: "v1.0",
		},
		{
			RuleID: "microshift-advent-x-10s-11-48t", Brand: "microSHIFT", CassetteSpec: "microshift_advent_x_10s_11_48t",
			DisplayName: "Advent X 山地 10速 11-48T", HintGroupsets: "Advent X",
			Speed: 10, MinCogTeeth: 11, MaxCogTeeth: 48, RecommendedFreehub: FreehubStandardHG,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG, "Shimano HG（山地）", noSpacer(), imageHGMountain, "山地 HG 塔基直接安装。")}, ImageSrc: imageHGMountain, MechanicalNotes: "11T 起跳，兼容标准 HG。", RuleVersion: "v1.0",
		},
		{
			RuleID: "sunshine-road-12s-11-34t", Brand: "SUNSHINE", CassetteSpec: "sunshine_12s_road_11_34t",
			DisplayName: "公路改装12速 11-30T 至 11-34T", HintGroupsets: "SUNSHINE / VG Sports",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 34, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11", noSpacer(), imageHG11Road, "可在老款 HG 轮组上直接运行。")}, ImageSrc: imageHG11Road, MechanicalNotes: "11T 起跳，免换 XDR。", RuleVersion: "v1.0",
		},
		{
			RuleID: "sunshine-mtb-12s-11-52t", Brand: "SUNSHINE", CassetteSpec: "sunshine_12s_mtb_11_52t",
			DisplayName: "山地改装12速 11-50T / 11-52T", HintGroupsets: "SUNSHINE / VG Sports",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 52, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11（公路轮组）", spacer(1.85, "若安装在公路长塔基上，需 1.85 mm 垫圈。"), imageHG11Road, "可在普通 HG 轮组上运行。")}, ImageSrc: imageHG11Road, MechanicalNotes: "山地宽度飞轮装公路长塔基需垫圈。", RuleVersion: "v1.0",
		},
		{
			RuleID: "ztto-hg-12s-11-34t", Brand: "ZTTO", CassetteSpec: "ztto_12s_hg_11_34t",
			DisplayName: "SLR 超轻 HG 12速 11-32T / 11-34T / 11-50T", HintGroupsets: "ZTTO SLR HG",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 50, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11", noSpacer(), imageHG11Road, "标准 HG-11 塔基。")}, ImageSrc: imageHG11Road, MechanicalNotes: "11T 起跳的 HG 飞轮。", RuleVersion: "v1.0",
		},
		{
			RuleID: "ztto-xd-12s-10-52t", Brand: "ZTTO", CassetteSpec: "ztto_12s_xd_10_52t",
			DisplayName: "SLR 超轻 XD 12速 9-50T / 10-52T", HintGroupsets: "ZTTO SLR XD",
			Speed: 12, MinCogTeeth: 9, MaxCogTeeth: 52, RecommendedFreehub: FreehubStandardXD,
			FitmentOptions: []FreehubFitmentOption{
				fitment(FreehubStandardXD, "SRAM XD", noSpacer(), imageXD, "XD 塔基直接安装。"),
				fitment(FreehubStandardXDR, "SRAM XDR", spacer(1.85, "XDR 塔基必须在底座加装 1.85 mm 垫圈。"), imageXDR, "XDR 比 XD 长 1.85 mm。"),
			}, ImageSrc: imageXD, MechanicalNotes: "9T/10T 起跳不能使用 HG。", RuleVersion: "v1.0",
		},
		{
			RuleID: "ztto-ms-12s-10-52t", Brand: "ZTTO", CassetteSpec: "ztto_12s_ms_10_52t",
			DisplayName: "SLR 超轻 Micro Spline 12速 10-51T / 10-52T", HintGroupsets: "ZTTO SLR MS",
			Speed: 12, MinCogTeeth: 10, MaxCogTeeth: 52, RecommendedFreehub: FreehubStandardMicroSpline,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardMicroSpline, "Shimano Micro Spline", noSpacer(), imageMicroSpline, "直接安装，无垫圈。")}, ImageSrc: imageMicroSpline, MechanicalNotes: "10T 起跳的 Micro Spline 飞轮。", RuleVersion: "v1.0",
		},
		{
			RuleID: "wheeltop-11-12s-11-34t", Brand: "Wheeltop", CassetteSpec: "wheeltop_11_12s_11_34t",
			DisplayName: "EDS TX 11/12速 11-32T / 11-34T", HintGroupsets: "Wheeltop EDS TX",
			Speed: 12, MinCogTeeth: 11, MaxCogTeeth: 34, RecommendedFreehub: FreehubStandardHG11,
			FitmentOptions: []FreehubFitmentOption{fitment(FreehubStandardHG11, "Shimano HG-11", noSpacer(), imageHG11Road, "直接安装，无垫圈。")}, ImageSrc: imageHG11Road, MechanicalNotes: "标配飞轮使用 Shimano HG 接口。", RuleVersion: "v1.0",
		},
	}

	// Keep provenance attached to every rule so API consumers, SSR pages, and
	// AI crawlers can verify the source family behind each compatibility claim.
	sourceRefsByBrand := map[string][]string{
		"SRAM":       {"https://www.sram.com/en/service"},
		"Shimano":    {"https://si.shimano.com/"},
		"Campagnolo": {"https://support.campagnolo.com/"},
		"L-TWOO":     {"https://ltwoo.com/"},
		"SENSAH":     {"https://sensah.com/"},
		"microSHIFT": {"https://microshift.com/"},
		"SUNSHINE":   {"https://www.sunshine-bike.com/"},
		"ZTTO":       {"https://ztto.com/"},
		"Wheeltop":   {"https://wheeltop.com/"},
	}
	for index := range rules {
		if refs := sourceRefsByBrand[rules[index].Brand]; len(refs) > 0 {
			rules[index].SourceRefs = append([]string(nil), refs...)
		}
	}
	return rules
}
