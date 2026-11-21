package ui

import (
	"github.com/manifoldco/promptui"
)

type ProductTypeSelectOption struct {
	Index              int
	Text               string
	SourceType         int
	AcceptProductTypes []string
	NeedSelectArticle  bool
	IsEnterpriseMode   bool
	IsUniversityMode   bool
}

var enterpriseProductTypeOptions = []ProductTypeSelectOption{
	{0, "训练营", 5, []string{"c44"}, true, true, false}, //custom source type, not use
}
var normalProductTypeOptions = []ProductTypeSelectOption{
	{0, "普通课程", 1, []string{"c1", "c3"}, true, false, false},
	{1, "每日一课", 2, []string{"d"}, false, false, false},
	{2, "公开课", 1, []string{"p35", "p29", "p30"}, true, false, false},
	{3, "大厂案例", 4, []string{"q"}, false, false, false},
	{4, "训练营", 5, []string{""}, true, false, true}, //custom source type, not use
	{5, "其他", 1, []string{"x", "c6"}, true, false, false},
}

func GetProductTypeSelectOptionByIndex(index int, isEnterprise bool) ProductTypeSelectOption {
	if isEnterprise {
		return enterpriseProductTypeOptions[index]
	}
	return normalProductTypeOptions[index]
}

func ProductTypeSelect(isEnterprise bool) (ProductTypeSelectOption, error) {

	productTypeOptions := []ProductTypeSelectOption{}

	if isEnterprise {
		productTypeOptions = enterpriseProductTypeOptions
	} else {
		productTypeOptions = normalProductTypeOptions
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ `>` | red }} {{ .Text | red }}",
		Inactive: "{{ .Text }}",
	}
	prompt := promptui.Select{
		Label:        "请选择想要下载的产品类型",
		Items:        productTypeOptions,
		Templates:    templates,
		Size:         len(productTypeOptions),
		HideSelected: true,
		Stdout:       NoBellStdout,
	}
	index, _, err := prompt.Run()
	if err != nil {
		return ProductTypeSelectOption{}, err
	}
	return productTypeOptions[index], nil
}

// IsUniversity checks if the product type is university product type
func (p *ProductTypeSelectOption) IsUniversity() bool {
	return p.IsUniversityMode
}
