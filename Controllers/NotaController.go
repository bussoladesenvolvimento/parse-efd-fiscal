package Controllers

import (
	"fmt"
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
	   Joins("join items on nota_fiscals.id=items.nota_fiscal_id").
	   Joins("left join emitentes  on emitentes.id = nota_fiscals.emitente_id").
	   Joins("left join destinatarios  on destinatarios.id = nota_fiscals.destinatario_id").
	   Find(&notas)


	color.Green("Fim popula xmls %s", time.Now())

	return &notas

}

func ExcelAddNota(notas *[]NotaFiscal.NotaFiscal, sheet *xlsx.Sheet) {

	nfs := *notas
	idnota := nfs[0].NNF
	count := 0
	for _, nf := range *notas {
		if nf.NNF != idnota {
			count = 0
			idnota = nf.NNF
		}
		fmt.Println("Nota : ", idnota)
		ExcelNota(sheet, nf, count)
		count++
		//for _, it := range nf.Itens {
		//	fmt.Println("Item: ", it.Descricao)
		//	ExcelNota(sheet, nf, it)
		//}
	}
}

func ExcelNota(sheet *xlsx.Sheet, nf NotaFiscal.NotaFiscal, i int) {

	fmt.Println("Total Itens da Nota: ", len(nf.Itens))
	fmt.Println("Item índice: ", i, " é ", nf.Itens[i].Descricao)
	menu := sheet.AddRow()

	//Nota Fiscal
	ColunaAdd(menu, nf.NNF)
	ColunaAdd(menu, nf.ChNFe)
	ColunaAdd(menu, nf.NatOp)
	ColunaAdd(menu, nf.IndPag)
	ColunaAdd(menu, nf.Mod)
	ColunaAdd(menu, nf.Serie)
	ColunaAdd(menu, "")
	ColunaAdd(menu, nf.TpNF)
	ColunaAdd(menu, nf.TpImp)
	ColunaAdd(menu, nf.TpEmis)
	ColunaAdd(menu, nf.CDV)
	ColunaAdd(menu, nf.TpAmb)
	ColunaAdd(menu, nf.FinNFe)
	ColunaAdd(menu, nf.ProcEmi)
	ColunaAdd(menu, "")
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
	ColunaAdd(menu, "")
	ColunaAdd(menu, nf.Destinatario.CNPJ)
	ColunaAdd(menu, nf.Destinatario.XNome)
	ColunaAdd(menu, nf.Destinatario.XLgr)
	ColunaAdd(menu, nf.Destinatario.Nro)
	ColunaAdd(menu, nf.Destinatario.XCpl)
	ColunaAdd(menu, nf.Destinatario.XBairro)
	ColunaAdd(menu, nf.Destinatario.CMun)
	ColunaAdd(menu, nf.Destinatario.XMun)
	ColunaAdd(menu, nf.Destinatario.Uf)
	ColunaAdd(menu, nf.Destinatario.Cep)
	ColunaAdd(menu, nf.Destinatario.CPais)
	ColunaAdd(menu, nf.Destinatario.XPais)
	ColunaAdd(menu, nf.Destinatario.Fone)
	ColunaAdd(menu, nf.Destinatario.Ie)
	ColunaAdd(menu, "")
	ColunaAdd(menu, nf.Itens[i].Codigo)
	ColunaAdd(menu, nf.Itens[i].Descricao)
	ColunaAdd(menu, nf.Itens[i].Unid)
	ColunaAdd(menu, fmt.Sprintf("%f", nf.Itens[i].Qtd))
	ColunaAdd(menu, fmt.Sprintf("%f", nf.Itens[i].VUnit))
	ColunaAdd(menu, fmt.Sprintf("%f", nf.Itens[i].VTotal))
	ColunaAdd(menu, nf.Itens[i].Cfop)
	ColunaAdd(menu, nf.Itens[i].DtEmit.String())
	ColunaAdd(menu, nf.Itens[i].Ean)
	ColunaAdd(menu, nf.Itens[i].Ncm)

}

func ExcelMenuNota(sheet *xlsx.Sheet) {
	menu := sheet.AddRow()

	//Nota Fiscal
	ColunaAdd(menu, "NNF")
	ColunaAdd(menu, "ChNFe")
	ColunaAdd(menu, "NatOp")
	ColunaAdd(menu, "IndPag")
	ColunaAdd(menu, "Mod")
	ColunaAdd(menu, "Serie")
	ColunaAdd(menu, "(EMISSAO)")
	ColunaAdd(menu, "TpNF")
	ColunaAdd(menu, "TpImp")
	ColunaAdd(menu, "TpEmis")
	ColunaAdd(menu, "CDV")
	ColunaAdd(menu, "TpAmb")
	ColunaAdd(menu, "FinNFe")
	ColunaAdd(menu, "ProcEmi")
	ColunaAdd(menu, "(EMITENTE)")
	ColunaAdd(menu, "CNPJ")
	ColunaAdd(menu, "XNome")
	ColunaAdd(menu, "XLgr")
	ColunaAdd(menu, "Nro")
	ColunaAdd(menu, "XCpl")
	ColunaAdd(menu, "XBairro")
	ColunaAdd(menu, "CMun")
	ColunaAdd(menu, "XMun")
	ColunaAdd(menu, "Uf")
	ColunaAdd(menu, "Cep")
	ColunaAdd(menu, "CPais")
	ColunaAdd(menu, "XPais")
	ColunaAdd(menu, "Fone")
	ColunaAdd(menu, "Ie")
	ColunaAdd(menu, "(DESTINATARIO)")
	ColunaAdd(menu, "CNPJ")
	ColunaAdd(menu, "XNome")
	ColunaAdd(menu, "XLgr")
	ColunaAdd(menu, "Nro")
	ColunaAdd(menu, "XCpl")
	ColunaAdd(menu, "XBairro")
	ColunaAdd(menu, "CMun")
	ColunaAdd(menu, "XMun")
	ColunaAdd(menu, "Uf")
	ColunaAdd(menu, "Cep")
	ColunaAdd(menu, "CPais")
	ColunaAdd(menu, "XPais")
	ColunaAdd(menu, "Fone")
	ColunaAdd(menu, "Ie")
	ColunaAdd(menu, "(ITEM)")
	ColunaAdd(menu,"Codigo")
	ColunaAdd(menu,"Descrição")
	ColunaAdd(menu,"Unid")
	ColunaAdd(menu, "Qtd")
	ColunaAdd(menu, "VUnit")
	ColunaAdd(menu, "VTotal")
	ColunaAdd(menu, "Cfop")
	ColunaAdd(menu, "DtEmit")
	ColunaAdd(menu, "Ean")
	ColunaAdd(menu, "Ncm")
}

