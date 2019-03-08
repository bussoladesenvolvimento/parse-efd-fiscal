package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bussoladesenvolvimento/parse-efd-fiscal/Controllers"
	"github.com/bussoladesenvolvimento/parse-efd-fiscal/Models/NotaFiscal"
	"github.com/bussoladesenvolvimento/parse-efd-fiscal/SpedDB"
	"github.com/bussoladesenvolvimento/parse-efd-fiscal/SpedRead"
	"github.com/bussoladesenvolvimento/parse-efd-fiscal/config"
	"github.com/bussoladesenvolvimento/parse-efd-fiscal/tools"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/tealeg/xlsx"
	"log"
	"net/http"
	"sync"
	"time"
)

var schema = flag.Bool("schema", false, "Recria as tabelas")
var importa = flag.Bool("importa", false, "Importa os xmls e speds ")
var ecf = flag.Bool("ecf", false, "Importa os ecfs ")
var inventario = flag.Bool("inventario", false, "Fazer processamento do inventario")
var anoInicial = flag.Int("anoInicial", 2012, "Ano inicial do processamento do inventário")
var anoFinal = flag.Int("anoFinal", 2019, "Ano inicial do processamento do inventário")
var excel = flag.Bool("excel", false, "Gera arquivo excel do inventario")
var excelNota = flag.Bool("excelNota", true, "Gera arquivo excel da nota")
var h010 = flag.Bool("h010", false, "Gera arquivo h010 e 0200 no layout sped para ser importado")


func init() {
	flag.Parse()
	cfg := new(config.Configurador)
	config.InicializaConfiguracoes(cfg)
}

func run(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error){

	ht, err := ParseHttpRequest(r)

	file, err := GetFile("file", ht)
	if err != nil {
		log.Println(err.Error())
		return nil, nil
	}

	fmt.Println(file.Name)


	dialect, err := config.Propriedades.ObterTexto("bd.dialect")
	conexao, err := config.Propriedades.ObterTexto("bd.conexao.mysql")
	digitos, err := config.Propriedades.ObterTexto("bd.digit.cod")
	db, err := gorm.Open(dialect, conexao)
	db.LogMode(true)
	//defer db.Close()
	if err != nil {
		fmt.Println("Falha ao abrir conexão. dialect=?, Linha de Conexao=?", dialect, conexao)
		return WriteStruct(NewUploadError(), http.StatusBadRequest)
	}

	if *schema {
		// Recria o Schema do banco de dados
		SpedDB.Schema(*db)
	}

	//if *ecf {
	//	// Lendo todos arquivos da pasta speds
	//	fmt.Println("Iniciando processamento ecf", time.Now())
	//	EcfRead.RecursiveEcfs("./ecfs", dialect, conexao, digitos)
	//	// Pega cada arquivo e ler linha a linha e envia para o banco de dados
	//	fmt.Println("Final processamento ecf", time.Now())
	//	var s string
	//	fmt.Scanln(&s)
	//}

	notas := []NotaFiscal.NotaFiscal{}
	if *importa {

		// Lendo todos arquivos da pasta speds
		fmt.Println("Iniciando processamento ", time.Now())
		SpedRead.RecursiveSpeds("./speds", &notas, dialect, conexao, digitos)
		// Pega cada arquivo e ler linha a linha e envia para o banco de dados
		fmt.Println("Final processamento ", time.Now())
		var s string
		fmt.Scanln(&s)
	}

	if *excelNota {


		notas := Controllers.PopularItens(*db)

		var file *xlsx.File
		var sheet *xlsx.Sheet
		var err error

		file = xlsx.NewFile()

		sheet, err = file.AddSheet(tools.PLANILHA_NOTA)
		if err != nil {
			fmt.Println(err)
			return WriteStruct(NewUploadError(), http.StatusBadRequest)
		}

		Controllers.ExcelMenuNota(sheet)
		Controllers.ExcelAddNota(notas, sheet)

		err = file.Save("NotasFiscais.xlsx")
		if err != nil {
			fmt.Println(err)
			return WriteStruct(NewUploadError(), http.StatusBadRequest)
		} else {
			fmt.Println("Arquivo de Analise Inventario Gerado com Sucesso!!!")
		}

	}

	if *inventario {
		//Recria tabela de inventário
		SpedDB.DropSchemaInventario(*db)
		SpedDB.CreateSchemaInventario(*db)

		// Processa o inventário
		fmt.Println("Inventario começou a processar as ?", time.Now())
		var wg sync.WaitGroup

		if *anoInicial == 0 || *anoFinal == 0 {
			fmt.Println("Favor informar o ano inicial que deseja processar. Exemplo -anoInicial=2011")
			return WriteStruct(NewUploadError(), http.StatusBadRequest)
		} else if *anoInicial <= 2011 || *anoFinal <= 2011 {
			fmt.Println("Favor informar um ano maior que 2011")
			return WriteStruct(NewUploadError(), http.StatusBadRequest)
		} else if *anoInicial <= 999 || *anoFinal <= 999 {
			fmt.Println("Favor informar o ano com 4 digitos. Exemplo 2017")
			return WriteStruct(NewUploadError(), http.StatusBadRequest)
		} else if *anoInicial > *anoFinal {
			fmt.Println("O ano inicial deve ser menor que o ano final")
		}

		wg.Add(2)
		go Controllers.ProcessarFatorConversao(*db, &wg)
		go Controllers.DeletarItensNotasCanceladas(*db, "2012-01-01", "2016-12-31", &wg)
		wg.Wait()

		wg.Add(2)
		go Controllers.PopularReg0200(*db, &wg)
		go Controllers.PopularItensXmls(*db, &wg)
		wg.Wait()

		wg.Add(3)
		go Controllers.PopularInventarios(*anoInicial, *anoFinal, &wg, *db)
		go Controllers.PopularEntradas(*anoInicial, *anoFinal, &wg, *db)
		go Controllers.PopularSaidas(*anoInicial, *anoFinal, &wg, *db)
		wg.Wait()

		// Quando finalizar todas essas deve rodar o processar diferencas
		Controllers.ProcessarDiferencas(*db)
		time.Sleep(90 * time.Second)
		fmt.Println(time.Now())
		color.Green("TERMINOU")
	}

	if *excel {
		var file *xlsx.File
		var sheet *xlsx.Sheet
		var err error

		file = xlsx.NewFile()

		sheet, err = file.AddSheet(tools.PLANILHA)
		if err != nil {
			fmt.Println(err)
			return WriteStruct(NewUploadError(), http.StatusBadRequest)
		}

		Controllers.ExcelMenu(sheet)
		Controllers.ExcelAdd(*db, sheet)

		err = file.Save("AnaliseInventario.xlsx")
		if err != nil {
			fmt.Println(err)
			return WriteStruct(NewUploadError(), http.StatusBadRequest)
		} else {
			fmt.Println("Arquivo de Analise Inventario Gerado com Sucesso!!!")
		}
	}

	if *h010 {

		if *anoInicial != 0 {
			//Controllers.CriarH010InvInicial(*ano, *db)
			//Controllers.CriarH010InvFinal(*ano, *db)
		} else {
			fmt.Println("Favor informar a tag ano. Exemplo: -ano=2016")
		}

	}

	resp, err := WriteStruct(&APIResponse{Success: true}, 200)

	if err != nil {
		log.Println(err.Error())
		return nil, nil
	}

	return resp, nil
}

func main() {

	lambda.Start(run)

}
