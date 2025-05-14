package justeat

import "github.com/WiiLink24/DemaeJustEat/demae"

var categoryTypes = map[demae.CategoryCode][]string{
	demae.Pizza:            {"pizza"},
	demae.Western:          {"hamburger", "pollo", "americano", "messicano", "american"},
	demae.FastFood:         {"panini", "hamburger", "friti", "pollo"},
	demae.Chinese:          {"cinese", "asianfusion"},
	demae.DrinksAndDessert: {"dolci", "gelato", "bevande"},
	demae.Curry:            {""},
	demae.Japanese:         {"asianfusion", "ramen"},
	demae.PartyFood:        {"kebab"},
	demae.Sushi:            {"sushi"},
}
