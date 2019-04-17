package NotaFiscal

import (
	"github.com/jinzhu/gorm"
	"time"
)

// Cadastro Itens referente aos campos da nota fiscal
type Item struct {
	gorm.Model
	Codigo       string    `gorm:"type:varchar(60)"`
	Ean          string    `gorm:"type:varchar(60)"`
	Descricao    string    `gorm:"type:varchar(100)"`
	Ncm          string    `gorm:"type:varchar(60)"`
	Cfop         string    `gorm:"type:varchar(60)"`
	Unid         string    `gorm:"type:varchar(60)"`
	Qtd          float64   `gorm:"type:decimal(19,3)"`
	VUnit        float64   `gorm:"type:decimal(19,3)"`
	VTotal       float64   `gorm:"type:decimal(19,3)"`
	DtEmit       time.Time `gorm:"type:date"`

	//ICMS
	IcmsModBC    int
	IcmsCST      int
	IcmsVBC      float64   `gorm:"type:decimal(19,3)"`
	IcmsPerc     float64   `gorm:"type:decimal(19,3)"`
	IcmsVal      float64   `gorm:"type:decimal(19,3)"`

	//PIS
	PisCST      int
	PisVBC      float64   `gorm:"type:decimal(19,3)"`
	PisPerc     float64   `gorm:"type:decimal(19,3)"`
	PisVal      float64   `gorm:"type:decimal(19,3)"`

	//COFINS
	CofinsCST     int
	CofinsVBC     float64   `gorm:"type:decimal(19,3)"`
	CofinsPerc    float64   `gorm:"type:decimal(19,3)"`
	CofinsVal     float64   `gorm:"type:decimal(19,3)"`

	NotaFiscalID uint
}
