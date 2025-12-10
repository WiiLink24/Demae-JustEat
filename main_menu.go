package main

import (
	"encoding/xml"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"strings"
)

func documentTemplate(r *Response) {
	r.AddKVWChildNode("container0", demae.KVField{
		XMLName: xml.Name{Local: "contents"},
		Value:   "TODO: Privacy Policy and Terms of Service",
	})
	r.AddKVWChildNode("container1", demae.KVField{
		XMLName: xml.Name{Local: "contents"},
		// Delivery success
		Value: "Enjoy your food!",
	})
	r.AddKVWChildNode("container2", demae.KVField{
		XMLName: xml.Name{Local: "contents"},
		// Delivery failure
		Value: "Contact WiiLink Support with your Wii Number",
	})
}

func categoryList(r *Response) {
	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId(), rdb)
	if err != nil {
		r.ReportError(err)
		return
	}

	categories, err := client.GetBareRestaurants()
	if err != nil {
		r.ReportError(err)
		return
	}

	r.MakeCategoryXMLs(categories)
	placeholder := demae.KVFieldWChildren{
		XMLName: xml.Name{Local: "Placeholder"},
		Value: []any{
			demae.KVField{
				XMLName: xml.Name{Local: "LargeCategoryName"},
				Value:   "Meal",
			},
			demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "CategoryList"},
				Value: []any{
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "TestingCategory"},
						Value: []any{
							demae.KVField{
								XMLName: xml.Name{Local: "CategoryCode"},
								Value:   "11",
							},
							demae.KVFieldWChildren{
								XMLName: xml.Name{Local: "ShopList"},
								Value: []any{
									demae.BasicShop{
										ShopCode:    demae.CDATA{Value: 0},
										HomeCode:    demae.CDATA{Value: 1},
										Name:        demae.CDATA{Value: "Test"},
										Catchphrase: demae.CDATA{Value: "A"},
										MinPrice:    demae.CDATA{Value: 1},
										Yoyaku:      demae.CDATA{Value: 1},
										Activate:    demae.CDATA{Value: "on"},
										WaitTime:    demae.CDATA{Value: 10},
										PaymentList: demae.KVFieldWChildren{
											XMLName: xml.Name{Local: "paymentList"},
											Value: []any{
												demae.KVField{
													XMLName: xml.Name{Local: "athing"},
													Value:   "Fox Card",
												},
											},
										},
										ShopStatus: demae.KVFieldWChildren{
											XMLName: xml.Name{Local: "shopStatus"},
											Value: []any{
												demae.KVFieldWChildren{
													XMLName: xml.Name{Local: "status"},
													Value: []any{
														demae.KVField{
															XMLName: xml.Name{Local: "isOpen"},
															Value:   1,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// It there is no nearby stores, we do not add the placeholder. This will tell the user there are no stores.
	if categories != nil && r.request.URL.Query().Get("action") != "webApi_shop_list" {
		r.AddCustomType(placeholder)
	}
}

func (r *Response) MakeCategoryXMLs(code []demae.CategoryCode) {
	for _, categoryCode := range code {
		r.AddCustomType(demae.KVFieldWChildren{
			XMLName: xml.Name{Local: string(categoryCode)},
			Value: []any{
				demae.KVField{
					XMLName: xml.Name{Local: "LargeCategoryName"},
					Value:   "Meal",
				},
				demae.KVFieldWChildren{
					XMLName: xml.Name{Local: "CategoryList"},
					Value: []any{
						demae.KVFieldWChildren{
							XMLName: xml.Name{Local: "TestingCategory"},
							Value: []any{
								demae.KVField{
									XMLName: xml.Name{Local: "CategoryCode"},
									Value:   categoryCode,
								},
								demae.KVFieldWChildren{
									XMLName: xml.Name{Local: "ShopList"},
									Value: []any{
										demae.BasicShop{
											ShopCode:    demae.CDATA{Value: 0},
											HomeCode:    demae.CDATA{Value: 1},
											Name:        demae.CDATA{Value: "Test"},
											Catchphrase: demae.CDATA{Value: "A"},
											MinPrice:    demae.CDATA{Value: 1},
											Yoyaku:      demae.CDATA{Value: 1},
											Activate:    demae.CDATA{Value: "on"},
											WaitTime:    demae.CDATA{Value: 10},
											PaymentList: demae.KVFieldWChildren{
												XMLName: xml.Name{Local: "paymentList"},
												Value: []any{
													demae.KVField{
														XMLName: xml.Name{Local: "athing"},
														Value:   "Fox Card",
													},
												},
											},
											ShopStatus: demae.KVFieldWChildren{
												XMLName: xml.Name{Local: "shopStatus"},
												Value: []any{
													demae.KVFieldWChildren{
														XMLName: xml.Name{Local: "status"},
														Value: []any{
															demae.KVField{
																XMLName: xml.Name{Local: "isOpen"},
																Value:   1,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
	}
}

func inquiryDone(r *Response) {
	// For our purposes, we will not be handling any restaurant requests.
	// However, the error endpoint uses this, so we must deal with that.
	// An error will never send a phone number, check for that first.
	if r.request.PostForm.Get("tel") != "" {
		return
	}

	shiftJisDecoder := func(str string) string {
		ret, _ := io.ReadAll(transform.NewReader(strings.NewReader(str), japanese.ShiftJIS.NewDecoder()))
		return string(ret)
	}

	// Now handle error.
	errorString := fmt.Sprintf(
		"An error has occured at on request %s\nError message: %s",
		shiftJisDecoder(r.request.PostForm.Get("requestType")),
		shiftJisDecoder(r.request.PostForm.Get("message")),
	)

	r.ReportError(fmt.Errorf(errorString))
}
