package setting

const WebsiteProfileGroup = "website_profile"

const (
	WebsiteProfileKeyEyebrow             = "eyebrow"
	WebsiteProfileKeyTitle               = "title"
	WebsiteProfileKeyLead                = "lead"
	WebsiteProfileKeyScope               = "scope"
	WebsiteProfileKeyAvatarURL           = "avatar_url"
	WebsiteProfileKeyAvatarLabel         = "avatar_label"
	WebsiteProfileKeyAvatarMark          = "avatar_mark"
	WebsiteProfileKeyProfileLabel        = "profile_label"
	WebsiteProfileKeyProfileRole         = "profile_role"
	WebsiteProfileKeyProfileContext      = "profile_context"
	WebsiteProfileKeyStatementEyebrow    = "statement_eyebrow"
	WebsiteProfileKeyStatementTitle      = "statement_title"
	WebsiteProfileKeyStatementParagraph1 = "statement_paragraph_1"
	WebsiteProfileKeyStatementParagraph2 = "statement_paragraph_2"
	WebsiteProfileKeyFactoryImageURL     = "factory_image_url"
	WebsiteProfileKeyFactoryImageAlt     = "factory_image_alt"
	WebsiteProfileKeyFactoryImageCaption = "factory_image_caption"
	WebsiteProfileKeyFactoryEyebrow      = "factory_eyebrow"
	WebsiteProfileKeyFactoryTitle        = "factory_title"
	WebsiteProfileKeyFactoryBody         = "factory_body"
	WebsiteProfileKeyFactoryCTA          = "factory_cta"
	WebsiteProfileKeyFactoryLink         = "factory_link"
)

type WebsiteProfileSettings struct {
	Locale              string `json:"locale"`
	Eyebrow             string `json:"eyebrow"`
	Title               string `json:"title"`
	Lead                string `json:"lead"`
	Scope               string `json:"scope"`
	AvatarURL           string `json:"avatar_url"`
	AvatarLabel         string `json:"avatar_label"`
	AvatarMark          string `json:"avatar_mark"`
	ProfileLabel        string `json:"profile_label"`
	ProfileRole         string `json:"profile_role"`
	ProfileContext      string `json:"profile_context"`
	StatementEyebrow    string `json:"statement_eyebrow"`
	StatementTitle      string `json:"statement_title"`
	StatementParagraph1 string `json:"statement_paragraph_1"`
	StatementParagraph2 string `json:"statement_paragraph_2"`
	FactoryImageURL     string `json:"factory_image_url"`
	FactoryImageAlt     string `json:"factory_image_alt"`
	FactoryImageCaption string `json:"factory_image_caption"`
	FactoryEyebrow      string `json:"factory_eyebrow"`
	FactoryTitle        string `json:"factory_title"`
	FactoryBody         string `json:"factory_body"`
	FactoryCTA          string `json:"factory_cta"`
	FactoryLink         string `json:"factory_link"`
}

type WebsiteProfileUpdateRequest struct {
	Locale              string `json:"locale"`
	Eyebrow             string `json:"eyebrow"`
	Title               string `json:"title"`
	Lead                string `json:"lead"`
	Scope               string `json:"scope"`
	AvatarURL           string `json:"avatar_url"`
	AvatarLabel         string `json:"avatar_label"`
	AvatarMark          string `json:"avatar_mark"`
	ProfileLabel        string `json:"profile_label"`
	ProfileRole         string `json:"profile_role"`
	ProfileContext      string `json:"profile_context"`
	StatementEyebrow    string `json:"statement_eyebrow"`
	StatementTitle      string `json:"statement_title"`
	StatementParagraph1 string `json:"statement_paragraph_1"`
	StatementParagraph2 string `json:"statement_paragraph_2"`
	FactoryImageURL     string `json:"factory_image_url"`
	FactoryImageAlt     string `json:"factory_image_alt"`
	FactoryImageCaption string `json:"factory_image_caption"`
	FactoryEyebrow      string `json:"factory_eyebrow"`
	FactoryTitle        string `json:"factory_title"`
	FactoryBody         string `json:"factory_body"`
	FactoryCTA          string `json:"factory_cta"`
	FactoryLink         string `json:"factory_link"`
}

func (request WebsiteProfileUpdateRequest) Settings() WebsiteProfileSettings {
	return WebsiteProfileSettings{
		Locale:              request.Locale,
		Eyebrow:             request.Eyebrow,
		Title:               request.Title,
		Lead:                request.Lead,
		Scope:               request.Scope,
		AvatarURL:           request.AvatarURL,
		AvatarLabel:         request.AvatarLabel,
		AvatarMark:          request.AvatarMark,
		ProfileLabel:        request.ProfileLabel,
		ProfileRole:         request.ProfileRole,
		ProfileContext:      request.ProfileContext,
		StatementEyebrow:    request.StatementEyebrow,
		StatementTitle:      request.StatementTitle,
		StatementParagraph1: request.StatementParagraph1,
		StatementParagraph2: request.StatementParagraph2,
		FactoryImageURL:     request.FactoryImageURL,
		FactoryImageAlt:     request.FactoryImageAlt,
		FactoryImageCaption: request.FactoryImageCaption,
		FactoryEyebrow:      request.FactoryEyebrow,
		FactoryTitle:        request.FactoryTitle,
		FactoryBody:         request.FactoryBody,
		FactoryCTA:          request.FactoryCTA,
		FactoryLink:         request.FactoryLink,
	}
}

func DefaultWebsiteProfileSettings(locale string) WebsiteProfileSettings {
	if locale == "zh_cn" {
		return WebsiteProfileSettings{
			Locale:              locale,
			Eyebrow:             "网站管理者 / 工厂成员",
			Title:               "我与这个网站",
			Lead:                "我负责这个网站的内容、结构和持续维护，也属于我们工厂正在做的事情。这个域名承载的是我的工作视角，而不是把我从工厂之外单独分离出来。",
			Scope:               "这个网站由我负责管理，也代表我们工厂的一部分工作",
			AvatarLabel:         "网站管理者头像位置",
			AvatarMark:          "我",
			ProfileLabel:        "网站管理者",
			ProfileRole:         "网站内容与方向",
			ProfileContext:      "我们工厂的一员",
			StatementEyebrow:    "为什么有这一页",
			StatementTitle:      "让网站背后的人被看见",
			StatementParagraph1: "这个域名属于我，但它表达的并不是一个脱离工厂的个人身份。相反，我希望用更接近个人的方式，说明我如何理解我们的工厂、产品和长期方向。",
			StatementParagraph2: "这里会记录网站背后的判断、正在推进的事情，以及我认为应该被准确表达的内容。它不是客服窗口，也不是单独成立的另一家公司，而是我们工厂工作中的一个管理和表达入口。",
			FactoryImageAlt:     "我们工厂的碳纤维手工铺层工序",
			FactoryImageCaption: "我负责表达的网站，来自我们真实的制造工作。",
			FactoryEyebrow:      "我们共同的工作",
			FactoryTitle:        "查看我们的工厂",
			FactoryBody:         "网站上的内容最终要回到真实的产品、制造、研发和质量控制。这里是我的视角，但它所指向的仍然是我们正在一起建设的工厂。",
			FactoryCTA:          "查看工厂与制造流程",
			FactoryLink:         "/company/about#factory",
		}
	}

	return WebsiteProfileSettings{
		Locale:              "en",
		Eyebrow:             "THE PERSON BEHIND THIS WEBSITE",
		Title:               "Me & This Website",
		Lead:                "I manage the content, structure, and ongoing maintenance of this website while remaining part of our factory and the work it represents. This domain carries my working perspective; it does not separate me from the factory.",
		Scope:               "Managed by me, grounded in the work of our factory",
		AvatarLabel:         "Website manager avatar placeholder",
		AvatarMark:          "ME",
		ProfileLabel:        "Website manager",
		ProfileRole:         "Content and direction",
		ProfileContext:      "Part of our factory",
		StatementEyebrow:    "WHY THIS PAGE EXISTS",
		StatementTitle:      "Let the person behind the site be visible",
		StatementParagraph1: "This domain belongs to me, but it does not describe a personal identity outside the factory. It gives me a more direct way to explain how I see our factory, our products, and the direction we are building toward.",
		StatementParagraph2: "This is where I can record the decisions behind the website, the work in progress, and the things I believe should be represented accurately. It is not a support desk or a separate company. It is one management and expression point within our factory work.",
		FactoryImageAlt:     "Carbon fiber hand layup work inside our factory",
		FactoryImageCaption: "The site I manage is grounded in our real manufacturing work.",
		FactoryEyebrow:      "THE WORK WE SHARE",
		FactoryTitle:        "See our factory",
		FactoryBody:         "The site should always lead back to real products, manufacturing, engineering, and quality control. This is my perspective, but it points to the factory we are building together.",
		FactoryCTA:          "View the factory and process",
		FactoryLink:         "/company/about#factory",
	}
}
