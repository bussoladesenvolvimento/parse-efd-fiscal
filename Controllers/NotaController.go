package Controllers

import (
	"github.com/bussoladesenvolvimento/parse-efd-fiscal/Models/NotaFiscal"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/tealeg/xlsx"
	"time"
)

func PopularItens(db gorm.DB)(*[]NotaFiscal.NotaFiscal) {
	color.Green("Comeco popula Itens da nota %s", time.Now())
	//result := make(map[string][]string)

	notas := []NotaFiscal.NotaFiscal{}

	db.Preload("Itens").
	   Preload("Emitente").
	   Preload("Destinatario").
	   Joins("join nota.items on nota.nota_fiscals.id=nota.items.nota_fiscal_id").
	   Joins("left join nota.emitentes  on nota.emitentes.id = nota.nota_fiscals.emitente_id").
	   Joins("left join nota.destinatarios  on nota.destinatarios.id = nota.nota_fiscals.destinatario_id").
	   Find(&notas)


	color.Green("Fim popula xmls %s", time.Now())

	return &notas

}

func ExcelAddNota(notas *[]NotaFiscal.NotaFiscal, sheet *xlsx.Sheet) {

	for _, nf := range *notas {
		ExcelNota(sheet, nf)
	}
}

func ExcelNota(sheet *xlsx.Sheet, nf NotaFiscal.NotaFiscal) {
	menu := sheet.AddRow()

	//Nota Fiscal
	ColunaAdd(menu, nf.NNF)
	ColunaAdd(menu, nf.ChNFe)
	ColunaAdd(menu, nf.NatOp)
	ColunaAdd(menu, nf.IndPag)
	ColunaAdd(menu, nf.Mod)
	ColunaAdd(menu, nf.Serie)
	ColunaAdd(menu, nf.DEmi)
	ColunaAdd(menu, nf.TpNF)
	ColunaAdd(menu, nf.TpImp)
	ColunaAdd(menu, nf.TpEmis)
	ColunaAdd(menu, nf.CDV)
	ColunaAdd(menu, nf.TpAmb)
	ColunaAdd(menu, nf.FinNFe)
	ColunaAdd(menu, nf.ProcEmi)
	ColunaAdd(menu, "Emitente")
	ColunaAdd(menu, nf.Emitente.CNPJ)
	ColunaAdd(menu, nf.Emitente.XNome)
	ColunaAdd(menu, nf.Emitente.XLgr)
	ColunaAdd(menu, nf.Emitente.Nro)
	ColunaAdd(menu, nf.Emitente.XCpl)
	ColunaAdd(menu, nf.Emitente.XBairro)
	ColunaAdd(menu, nf.Emitente.CMun)
	ColunaAdd(menu, nf.Emitente.XMun)
	ColunaAdd(menu, nf.Emitente.Uf)
	ColunaAdd(menu, nf.Emitente.Cep)
	ColunaAdd(menu, nf.Emitente.CPais)
	ColunaAdd(menu, nf.Emitente.XPais)
	ColunaAdd(menu, nf.Emitente.Fone)
	ColunaAdd(menu, nf.Emitente.Ie)

}

func ExcelMenuNota(sheet *xlsx.Sheet) {
	menu := sheet.AddRow()

	//Nota Fiscal
	ColunaAdd(menu, "NNF")

}

