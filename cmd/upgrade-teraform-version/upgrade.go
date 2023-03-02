package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func main() {
	files, _ := os.ReadDir(".")
	for _, file := range files {
		if file.Name() == "versions.tf" {
			data, err := os.ReadFile("versions.tf")
			if err != nil {
				fmt.Println(err)
			}

			f, diags := hclwrite.ParseConfig(data, file.Name(), hcl.Pos{
				Line:   1,
				Column: 1,
			})

			if diags.HasErrors() {
				fmt.Println(diags)
			}
			blocks := f.Body().Blocks()
			for _, block := range blocks {
				blockBody := block.Body()

				if blockBody.Attributes()["required_version"] == nil {
					continue
				}
				expr := blockBody.Attributes()["required_version"].Expr()
				exprTokens := expr.BuildTokens(nil)

				var valueTokens hclwrite.Tokens
				valueTokens = append(valueTokens, exprTokens...)

				blockBody.SetAttributeValue("required_version", cty.StringVal(">= 1.2.5"))

				err = os.WriteFile(file.Name(), f.Bytes(), 0o644)
				if err != nil {
					fmt.Println(err)
				}

				color.Blue("Updated %s", file.Name())
			}
		}
	}

}
