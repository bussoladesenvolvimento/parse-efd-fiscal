package Bloco0

import (
	"github.com/bussoladesenvolvimento/parse-efd-fiscal/tools"
	"github.com/jinzhu/gorm"
	"time"
)

// Estrutura criada usando layout Guia Prático EFD-ICMS/IPI – Versão 2.0.20 Atualização: 07/12/2016
// Estrutura modelo do banco de dados Registro 0190
type Reg0190 struct {
	gorm.Model
	Reg   string `gorm:"type:varchar(4)"`
	Unid  string `gorm:"type:varchar(6)"`
	Descr string
	DtIni time.Time `gorm:"type:date"`
	DtFin time.Time `gorm:"type:date"`
	Cnpj  string    `gorm:"type:varchar(14)"`
}

func (Reg0190) TableName() string {
	return "reg_0190"
}

type iReg0190 interface {
	GetReg0190() Reg0190
}

// Implementando Interface do Sped Reg0190
type Reg0190Sped struct {
	Ln      []string
	Reg0000 Reg0000
}

func (s Reg0190Sped) GetReg0190() Reg0190 {
	reg0190 := Reg0190{
		Reg:   s.Ln[1],
		Unid:  s.Ln[2],
		Descr: s.Ln[3],
		DtIni: s.Reg0000.DtIni,
		DtFin: s.Reg0000.DtFin,
		Cnpj:  s.Reg0000.Cnpj,
	}
	return reg0190
}

type Reg0190Xml struct {
	Data string
}

func (x Reg0190Xml) GetReg0190() Reg0190 {
	reg0190 := Reg0190{
		Reg:   "0190",
		Unid:  x.Data,
		Descr: "Importado Xml",
		DtIni: tools.ConvertDataNull(),
		DtFin: tools.ConvertDataNull(),
		Cnpj:  "",
	}
	return reg0190
}

func CreateReg0190(read iReg0190) Reg0190 {
	return read.GetReg0190()
}
