package justeat

import "github.com/WiiLink24/DemaeJustEat/demae"

var categoryTypes = map[demae.CategoryCode][]string{
	demae.Pizza:            {"pizza"},
	demae.Western:          {"hamburger", "pollo", "americano", "messicano", "american", "sandwiches"},
	demae.FastFood:         {"panini", "hamburger", "friti", "pollo", "chicken"},
	demae.Chinese:          {"cinese", "asianfusion", "asian", "chinese"},
	demae.DrinksAndDessert: {"dolci", "gelato", "bevande", "desserts", "coffee", "cakes", "drinks"},
	demae.Curry:            {"curry", "indian"},
	demae.Japanese:         {"asianfusion", "ramen", "japanese"},
	demae.PartyFood:        {"kebab"},
	demae.Sushi:            {"sushi"},
}
