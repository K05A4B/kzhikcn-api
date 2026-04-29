package parser

import (
	"fmt"
	"testing"
)

func TestLexerTokenize(t *testing.T) {
	dataSet := []string{
		"(age > 18 & city = \"New York\") | (score = 0 & height < 180) | (country = \"US\" & score = 0)",
		"age != 0 & index=`bbb | role=superAdmin & auth end=\"aaa\"",
		`age>18 & city="New York" &  (height<180) & !(ip_location ~ "US%" | score=null)`,
		`title ~ "%页面"`,
	}

	for _, data := range dataSet {
		t.Run(data, func(t *testing.T) {
			token, err := Tokenize(data)
			if err != nil {
				t.Error(err)
				return
			}

			fmt.Println(token)
		})
	}
}

func TestLexerHardTokenize(t *testing.T) {
	dataSet := []string{
		`   (   ( age    >=   25  &   city    =     "San    Francisco"  )  
   |   (  salary   <   50000    &   employed    =  true   )   )    
   & ( role  !=   "manager"   |   department   =  "IT"   )    
   |   ( height  >=    180    &   weight    <=    75  )    
   & ( name   =   "John   Doe"   |   married  =   false  )
`,
	}

	for _, data := range dataSet {
		token, err := Tokenize(data)
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println("+++++++++++++++++++++++++++++++++++++++")
		fmt.Println(token)
	}
}
