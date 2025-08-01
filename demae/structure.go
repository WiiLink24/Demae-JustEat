package demae

import (
	"encoding/xml"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCConfig struct {
	XMLName      xml.Name `xml:"oidc"`
	ClientID     string   `xml:"clientID"`
	ClientSecret string   `xml:"clientSecret"`
	RedirectURL  string   `xml:"redirectURL"`
	Scopes       []string `xml:"scopes"`
	Provider     string   `xml:"provider"`
}

type Config struct {
	XMLName        xml.Name   `xml:"Config"`
	DemaeAddress   string     `xml:"DemaeAddress"`
	JustEatAddress string     `xml:"JustEatAddress"`
	SQLAddress     string     `xml:"SQLAddress"`
	SQLUser        string     `xml:"SQLUser"`
	SQLPass        string     `xml:"SQLPass"`
	SQLDB          string     `xml:"SQLDB"`
	ErrorWebhook   string     `xml:"ErrorWebhook"`
	OrderWebhook   string     `xml:"OrderWebhook"`
	SentryDSN      string     `xml:"SentryDSN"`
	OIDCConfig     OIDCConfig `xml:"oidc"`
	IsDebug        bool       `xml:"IsDebug"`
}

type AppAuthConfig struct {
	OAuth2Config *oauth2.Config
	Provider     *oidc.Provider
}

// KVField represents an individual node in form of <XMLName>Contents</XMLName>.
type KVField struct {
	XMLName xml.Name
	Value   any `xml:",cdata"`
}

// KVFieldWChildren represents an individual node in form of
/*
<XMLName>
	<Child>
		...
	</Child>
</XMLName>
*/
type KVFieldWChildren struct {
	XMLName xml.Name
	Value   []any
}

type CDATA struct {
	Value any `xml:",cdata"`
}

type AreaNames struct {
	XMLName  xml.Name `xml:"areaPlace"`
	AreaName CDATA    `xml:"areaName"`
	AreaCode CDATA    `xml:"areaCode"`
}

type Area struct {
	XMLName    xml.Name `xml:"areaPlace"`
	AreaName   CDATA    `xml:"areaName"`
	AreaCode   CDATA    `xml:"areaCode"`
	IsNextArea CDATA    `xml:"isNextArea"`
	Display    CDATA    `xml:"display"`
	Kanji1     CDATA    `xml:"kanji1"`
	Kanji2     CDATA    `xml:"kanji2"`
	Kanji3     CDATA    `xml:"kanji3"`
	Kanji4     CDATA    `xml:"kanji4"`
}

type AreaShopInfo struct {
	XMLName  xml.Name `xml:"Shop"`
	ShopCode CDATA    `xml:"shopCode"`
}

type BasicShop struct {
	XMLName     xml.Name `xml:"Shop"`
	ShopCode    CDATA    `xml:"shopCode"`
	HomeCode    CDATA    `xml:"homeCode"`
	Name        CDATA    `xml:"name"`
	Catchphrase CDATA    `xml:"catchphrase"`
	MinPrice    CDATA    `xml:"minPrice"`
	Yoyaku      CDATA    `xml:"yoyaku"`
	Activate    CDATA    `xml:"activate"`
	WaitTime    CDATA    `xml:"waitTime"`
	PaymentList KVFieldWChildren
	ShopStatus  KVFieldWChildren
}

type ShopOne struct {
	XMLName             xml.Name `xml:"response"`
	CategoryCode        CDATA    `xml:"categoryCode"`
	Address             CDATA    `xml:"address"`
	Information         CDATA    `xml:"information"`
	Attention           CDATA    `xml:"attention"`
	Amenity             CDATA    `xml:"amenity"`
	MenuListCode        CDATA    `xml:"menuListCode"`
	Activate            CDATA    `xml:"activate"`
	WaitTime            CDATA    `xml:"waitTime"`
	TimeOrder           CDATA    `xml:"timeorder"`
	Tel                 CDATA    `xml:"tel"`
	YoyakuMinDate       CDATA    `xml:"yoyakuMinDate"`
	YoyakuMaxDate       CDATA    `xml:"yoyakuMaxDate"`
	PaymentList         KVFieldWChildren
	ShopStatus          ShopStatus       `xml:"shopStatus"`
	RecommendedItemList KVFieldWChildren `xml:"recommendItemList"`
}

type ShopStatus struct {
	Hours    KVFieldWChildren
	Interval CDATA `xml:"interval"`
	Holiday  CDATA `xml:"holiday"`
}

type NestedItem struct {
	XMLName xml.Name
	Name    CDATA `xml:"name"`
	Item    Item
}
type Item struct {
	XMLName    xml.Name
	MenuCode   CDATA             `xml:"menuCode"`
	ItemCode   CDATA             `xml:"itemCode"`
	Name       CDATA             `xml:"name"`
	Price      CDATA             `xml:"price"`
	Info       CDATA             `xml:"info"`
	Size       *CDATA            `xml:"size"`
	IsSelected *CDATA            `xml:"isSelected"`
	Image      CDATA             `xml:"image"`
	IsSoldout  CDATA             `xml:"isSoldout"`
	SizeList   *KVFieldWChildren `xml:"sizeList"`
}

type ItemSize struct {
	XMLName   xml.Name
	ItemCode  CDATA `xml:"itemCode"`
	Size      CDATA `xml:"size"`
	Price     CDATA `xml:"price"`
	IsSoldout CDATA `xml:"isSoldout"`
}

type Menu struct {
	XMLName       xml.Name
	MenuCode      CDATA `xml:"menuCode"`
	LinkTitle     CDATA `xml:"linkTitle"`
	EnabledLink   CDATA `xml:"enabledLink"`
	Name          CDATA `xml:"name"`
	Info          CDATA `xml:"info"`
	SetNum        CDATA `xml:"setNum"`
	LunchMenuList struct {
		IsLunchTimeMenu CDATA `xml:"isLunchTimeMenu"`
		Hour            KVFieldWChildren
		IsOpen          CDATA `xml:"isOpen"`
		Message         CDATA `xml:"message"`
	} `xml:"lunchMenuList"`
}

type ItemOne struct {
	XMLName xml.Name
	Info    CDATA            `xml:"info"`
	Code    CDATA            `xml:"code"`
	Type    CDATA            `xml:"type"`
	Name    CDATA            `xml:"name"`
	List    KVFieldWChildren `xml:"list"`
}

type BasketItem struct {
	XMLName       xml.Name
	BasketNo      CDATA            `xml:"basketNo"`
	MenuCode      CDATA            `xml:"menuCode"`
	ItemCode      CDATA            `xml:"itemCode"`
	Name          CDATA            `xml:"name"`
	Price         CDATA            `xml:"price"`
	Size          CDATA            `xml:"size"`
	IsSoldout     CDATA            `xml:"isSoldout"`
	Quantity      CDATA            `xml:"quantity"`
	SubTotalPrice CDATA            `xml:"subTotalPrice"`
	Menu          KVFieldWChildren `xml:"Menu"`
	OptionList    KVFieldWChildren `xml:"optionList"`
}
