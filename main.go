package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/clbanning/mxj"
	"io"
	"io/ioutil"
	"mime"
	"net/textproto"
	"net/url"
	"os"
	"reflect"
	"strconv"

	//"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
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
	"path/filepath"
	"sync"
	"time"
)

var schema = flag.Bool("schema", true, "Recria as tabelas")
var importa = flag.Bool("importa", true, "Importa os xmls e speds ")
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

func createSchema()(*gorm.DB, error){
	dialect, err := config.Propriedades.ObterTexto("bd.dialect")
	conexao, err := config.Propriedades.ObterTexto("bd.conexao.mysql")


	db, err := gorm.Open(dialect, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", "userdiniz", "apmGj]Jj]]Jo", "dbdiniz.cjayksip8ytz.us-east-1.rds.amazonaws.com", "3306", "dbdiniz"))
	db.LogMode(true)
	//defer db.Close()
	if err != nil {
		fmt.Println("Falha ao abrir conexão. dialect=?, Linha de Conexao=?", dialect, conexao)
		return  nil, err
	}

	if *schema {
		// Recria o Schema do banco de dados
		SpedDB.Schema(*db)
	}

	return db, nil
}




func runTeste(filename string) (){

	resrouce := "./speds/" + filename
	file, err := os.Open(resrouce) // For read access.
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	buf := bytes.Buffer{}
	if _, err := io.Copy(&buf, file); err != nil {
		log.Fatal(err)
	}

	digitosCodigo, err := config.Propriedades.ObterTexto("bd.digit.cod")
	xmlFile := buf.Bytes()

	digitosCodigo2 := tools.ConvInt(digitosCodigo)
	// Teste de lista produtos
	//xmlFile, err := ioutil.ReadFile(xml)
	//notas := [] NotaFiscal.NotaFiscal{}
	//if err := xml.NewDecoder(file).Decode(&notas); err != nil {
	//	log.Fatal(err)
	//	return
	//}

	reader := tools.ConvXmlBye(xmlFile)
	tools.CheckErr(err)
	nfe, errOpenXml := mxj.NewMapXml(xmlFile)
	tools.CheckErr(errOpenXml)
	// Preenchendo o header da nfe
	nNf := reader("ide", "nNF")
	chnfe := reader("infProt", "chNFe")
	natOp := reader("ide", "natOp")
	indPag := reader("ide", "indPag")
	mod := reader("ide", "mod")
	serie := reader("ide", "serie")
	dEmit := reader("ide", "dEmi")
	if dEmit == "" {
		dhEmit := reader("ide", "dhEmi")
		dEmit = dhEmit
	}
	tpNf := reader("ide", "tpNF")
	tpImp := reader("ide", "tpImp")
	tpEmis := reader("ide", "tpEmis")
	cdv := reader("ide", "cDV")
	tpAmb := reader("ide", "tpAmb")
	finNFe := reader("ide", "finNFe")
	procEmi := reader("ide", "procEmi")

	// Preenchendo itens
	codigo, err := nfe.ValuesForKey("cProd")
	ean, err := nfe.ValuesForKey("cEAN")
	descricao, err := nfe.ValuesForKey("xProd")
	ncm, err := nfe.ValuesForKey("NCM")
	cfop, err := nfe.ValuesForKey("CFOP")
	unid, err := nfe.ValuesForKey("uCom")
	qtd, err := nfe.ValuesForKey("qCom")
	vUnit, err := nfe.ValuesForKey("vUnCom")
	vTotal, err := nfe.ValuesForKey("vProd")

	icmsTotal, err := nfe.ValuesForKey("ICMSTot")
	vIcmsTotal := reflect.ValueOf(icmsTotal[0])
	fmt.Println(vIcmsTotal)

	vBCi := ""
	vICMS:= ""
	vICMSDeson:=""
	vBCST:=""
	vST:=""
	vProd:=""
	vFrete:=""
	vSeg:=""
	vDesc:=""
	vII:=""
	vIPI:=""
	vPIS:=""
	vCOFINS:=""
	vOutro:=""
	vNF:=""
	vTotTrib:=""
	for _, k := range vIcmsTotal.MapKeys() {

		fmt.Println(k.Interface())
		fmt.Println(vIcmsTotal)
		switch k.Interface() {
		case "vBC":
			value :=  vIcmsTotal.MapIndex(k)
			vBCi = fmt.Sprintf("%s", value)
		case "vICMS":
			value :=  vIcmsTotal.MapIndex(k)
			vICMS = fmt.Sprintf("%s", value)
		case "vICMSDeson":
			value :=  vIcmsTotal.MapIndex(k)
			vICMSDeson = fmt.Sprintf("%s", value)
		case "vBCST":
			value :=  vIcmsTotal.MapIndex(k)
			vBCST = fmt.Sprintf("%s", value)
		case "vST":
			value :=  vIcmsTotal.MapIndex(k)
			vICMS = fmt.Sprintf("%s", value)
		case "vProd":
			value :=  vIcmsTotal.MapIndex(k)
			vProd = fmt.Sprintf("%s", value)
		case "vFrete":
			value :=  vIcmsTotal.MapIndex(k)
			vFrete = fmt.Sprintf("%s", value)
		case "vSeg":
			value :=  vIcmsTotal.MapIndex(k)
			vSeg =fmt.Sprintf("%s", value)
		case "vDesc":
			value :=  vIcmsTotal.MapIndex(k)
			vDesc = fmt.Sprintf("%s", value)
		case "vII":
			value :=  vIcmsTotal.MapIndex(k)
			vII = fmt.Sprintf("%s", value)
		case "vIPI":
			value :=  vIcmsTotal.MapIndex(k)
			vIPI = fmt.Sprintf("%s", value)
		case "vPIS":
			value :=  vIcmsTotal.MapIndex(k)
			vPIS = fmt.Sprintf("%s", value)
		case "vCOFINS":
			value :=  vIcmsTotal.MapIndex(k)
			vCOFINS = fmt.Sprintf("%s", value)
		case "vOutro":
			value :=  vIcmsTotal.MapIndex(k)
			vOutro = fmt.Sprintf("%s", value)
		case "vNF":
			value :=  vIcmsTotal.MapIndex(k)
			vNF = fmt.Sprintf("%s", value)
		case "vTotTrib":
			value :=  vIcmsTotal.MapIndex(k)
			vTotTrib = fmt.Sprintf("%s", value)
		}

	}


	//vICMSi :=  mIcmsTotal.MapKeys()[1]
	//vICMSDesoni :=  mIcmsTotal.MapKeys()[2]
	//vBCSTi :=  mIcmsTotal.MapKeys()[3]
	//vSTi := mIcmsTotal.MapKeys()[4]
	//vProdi :=  mIcmsTotal.MapKeys()[5]
	//vFretei := vFrete[0].(string)
	//vSegi := vSeg[0].(string)
	//vDesci := vDesc[0].(string)
	//vIIi := vII[0].(string)
	//vIPIi := vIPI[0].(string)
	//vPISi := vPIS[0].(string)
	//vCOFINSi := vCOFINS[0].(string)
	//vOutroi := vOutro[0].(string)
	//vNFi := vNF[0].(string)
	//vTotTribi := vTotTrib[0].(string)

	imposto, err := nfe.ValuesForKey("imposto")
	//iCMSTot, err := nfe.ValuesForKey("ICMSTot")
	//fmt.Println(iCMSTot)

	// Preenchendo Destinatario
	cnpj := reader("dest", "CNPJ")
	xNome := reader("dest", "xNome")
	xLgr := reader("enderDest", "xLgr")
	nro := reader("enderDest", "nro")
	xCpl := reader("enderDest", "xCpl")
	xBairro := reader("enderDest", "xBairro")
	cMun := reader("enderDest", "cMun")
	xMun := reader("enderDest", "xMun")
	uf := reader("enderDest", "UF")
	cep := reader("enderDest", "CEP")
	cPais := reader("enderDest", "cPais")
	xPais := reader("enderDest", "xPais")
	fone := reader("enderDest", "fone")
	ie := reader("dest", "IE")
	// Preenchendo Emitente
	cnpje := reader("emit", "CNPJ")
	xNomee := reader("emit", "xNome")
	xLgre := reader("enderEmit", "xLgr")
	nroe := reader("enderEmit", "nro")
	xCple := reader("enderEmit", "xCpl")
	xBairroe := reader("enderEmit", "xBairro")
	cMune := reader("enderEmit", "cMun")
	xMune := reader("enderEmit", "xMun")
	ufe := reader("enderEmit", "UF")
	cepe := reader("enderEmit", "CEP")
	cPaise := reader("enderEmit", "cPais")
	xPaise := reader("enderEmit", "xPais")
	fonee := reader("enderEmit", "fone")
	iee := reader("emit", "IE")

	destinatario := NotaFiscal.Destinatario{
		CNPJ:    cnpj,
		XNome:   xNome,
		XLgr:    xLgr,
		Nro:     nro,
		XCpl:    xCpl,
		XBairro: xBairro,
		CMun:    cMun,
		XMun:    xMun,
		Uf:      uf,
		Cep:     cep,
		CPais:   cPais,
		XPais:   xPais,
		Fone:    fone,
		Ie:      ie,
	}

	emitentede := NotaFiscal.Emitente{
		CNPJ:    cnpje,
		XNome:   xNomee,
		XLgr:    xLgre,
		Nro:     nroe,
		XCpl:    xCple,
		XBairro: xBairroe,
		CMun:    cMune,
		XMun:    xMune,
		Uf:      ufe,
		Cep:     cepe,
		CPais:   cPaise,
		XPais:   xPaise,
		Fone:    fonee,
		Ie:      iee,
	}

	var itens []NotaFiscal.Item

	for i, _ := range codigo {
		i2 := i + 1
		codigoi := tools.AdicionaDigitosCodigo(codigo[i].(string), digitosCodigo2)
		eani := ean[i].(string)
		descricaoi := descricao[i].(string)
		ncmi := ncm[i].(string)
		cfopi := cfop[i].(string)
		unidi := unid[i].(string)
		qtdi := qtd[i].(string)
		vuniti := vUnit[i].(string)
		vtotali := vTotal[i2].(string)

		pCST  := ""
		pvBC  := ""
		pPIS  := ""
		pvPIS := ""

		impostoprod := imposto[i]
		v := reflect.ValueOf(impostoprod)
		fmt.Printf("Type: %v\n", v)     // map[]
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				strct := v.MapIndex(key)
				fmt.Println(key.Interface(), strct.Interface())

				if key.Interface() == "PIS" {
					teste:= reflect.ValueOf(strct.Interface())
					for _, kk := range teste.MapKeys() {
						pISAliq := teste.MapIndex(kk)
						mPis := reflect.ValueOf(pISAliq.Interface())
						for _, kkk := range mPis.MapKeys() {

							fmt.Println(kkk.Interface())
							switch kkk.Interface() {
							case "CST":
								value :=  mPis.MapIndex(kkk)
								pCST = fmt.Sprintf("%s", value)
							case "vBC":
								value :=  mPis.MapIndex(kkk)
								pvBC = fmt.Sprintf("%s", value)
							case "pPIS":
								value :=  mPis.MapIndex(kkk)
								pPIS = fmt.Sprintf("%s", value)
							case "vPIS":
								value :=  mPis.MapIndex(kkk)
								pvPIS = fmt.Sprintf("%s", value)
							}

						}
					}
				}


			}
		}

		Item := NotaFiscal.Item{
			Codigo:    codigoi,
			Ean:       eani,
			Descricao: descricaoi,
			Ncm:       ncmi,
			Cfop:      cfopi,
			Unid:      unidi,
			Qtd:       tools.ConvFloat(qtdi),
			VUnit:     tools.ConvFloat(vuniti),
			VTotal:    tools.ConvFloat(vtotali),
			DtEmit:    tools.ConvertDataXml(dEmit),
			PisCST:    tools.ConvInt(pCST),
			PisVBC:    tools.ConvFloat(pvBC),
			PisVal:    tools.ConvFloat(pvPIS),
			PisPerc:   tools.ConvFloat(pPIS),
		}
		itens = append(itens, Item)
	}

	notafiscal := NotaFiscal.NotaFiscal{
		NNF:          nNf,
		ChNFe:        chnfe,
		NatOp:        natOp,
		IndPag:       indPag,
		Mod:          mod,
		Serie:        serie,
		DEmi:         tools.ConvertDataXml(dEmit),
		TpNF:         tpNf,
		TpImp:        tpImp,
		TpEmis:       tpEmis,
		CDV:          cdv,
		TpAmb:        tpAmb,
		FinNFe:       finNFe,
		ProcEmi:      procEmi,
		Emitente:     emitentede,
		Destinatario: destinatario,
		ICMSTotVBC:   tools.ConvFloat(fmt.Sprintf("%s", vBCi)),
		ICMSTotVICMS: tools.ConvFloat(fmt.Sprintf("%s", vICMS)),
		ICMSTotVICMSDeson: tools.ConvFloat(fmt.Sprintf("%s", vICMSDeson)),
		ICMSTotVBCST: tools.ConvFloat(fmt.Sprintf("%s", vBCST)),
		ICMSTotVST: tools.ConvFloat(fmt.Sprintf("%s", vST)),
		ICMSTotVProd: tools.ConvFloat(fmt.Sprintf("%s", vProd)),
		ICMSTotVFrete: tools.ConvFloat(fmt.Sprintf("%s", vFrete)),
		ICMSTotVSeg: tools.ConvFloat(fmt.Sprintf("%s", vSeg)),
		ICMSTotVDesc: tools.ConvFloat(fmt.Sprintf("%s", vDesc)),
		ICMSTotVII: tools.ConvFloat(fmt.Sprintf("%s", vII)),
		ICMSTotVIPI: tools.ConvFloat(fmt.Sprintf("%s", vIPI)),
		ICMSTotVPIS: tools.ConvFloat(fmt.Sprintf("%s", vPIS)),
		ICMSTotVCOFINS: tools.ConvFloat(fmt.Sprintf("%s", vCOFINS)),
		ICMSTotVOutro: tools.ConvFloat(fmt.Sprintf("%s", vOutro)),
		ICMSTotVNF: tools.ConvFloat(fmt.Sprintf("%s", vNF)),
		ICMSTotVTotTrib: tools.ConvFloat(fmt.Sprintf("%s", vTotTrib)),
		Itens:        itens,
	}

	fmt.Println("Nota Fiscal: ", notafiscal)

}



func run(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error){

	db, err := createSchema()
	if err != nil {
		fmt.Println("Falha ao abrir conexão.")
		return WriteStruct(NewDbError(), http.StatusBadRequest)
	}

	ht, err := ParseHttpRequest(r)
	err = ht.ParseMultipartForm(ht.ContentLength)                     // Parses the request body
	if err != nil {
		log.Println(err.Error())
		return nil, nil
	}

	formdata := ht.MultipartForm
	files := formdata.File["multiplefiles"]
	for i, _ := range files {
		file, err := files[i].Open()
		fmt.Println("File ", file)
		defer file.Close()
		if err != nil {
			log.Println(err.Error())
			return nil, nil
		}

		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, file); err != nil {
			return nil, err
		}

		fmt.Println("Filename: ", files[i].Filename)
		fmt.Println("Header: ", files[i].Header)
		newFile := &File{Name: files[i].Filename,Content:buf.Bytes(),Header:files[i].Header}


		digitos, err := config.Propriedades.ObterTexto("bd.digit.cod")
		dialect, err := config.Propriedades.ObterTexto("bd.dialect")
		conexao, err := config.Propriedades.ObterTexto("bd.conexao.mysql")
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
			ext := filepath.Ext(newFile.Name)

			if ext == ".xml" || ext == ".XML" {

				SpedRead.InsertXml(&notas, newFile.Content, dialect, conexao, digitos)

				//go InsertXml(file, dialect, conexao, digitosCodigo)
			}

			//SpedRead.RecursiveSpeds("./speds", &notas, dialect, conexao, digitos)
			// Pega cada arquivo e ler linha a linha e envia para o banco de dados
			fmt.Println("Final processamento ", time.Now())
		}

		fmt.Println("Filename: ",  files[i].Filename)

	}
	//fmt.Println("ponto 3")
	//for i := range ht.Form {
	//	fmt.Println("ponto 4")
	//	ht.
	//	file, err := GetFile(i, ht)
	//	fmt.Println("ponto 5")
	//	fmt.Println("i: ", i)
	//	fmt.Println("ht: ", ht)
	//	fmt.Println("file: ", file)
	//	if err != nil {
	//		log.Println(err.Error())
	//		return nil, nil
	//	}
	//	fmt.Println("file ", file)
	//	fmt.Println("get key ", i)
    //    fmt.Println(ht.Form.Get(i))
	//}
	//fmt.Println("ponto 6")
	//files, fheader,err := ht.FormFile("attachments")
	//if err!=nil{
	//	fmt.Println("deu erro")
	//	fmt.Println(err)
	//	return nil,err
	//}
	//
	//if files != nil {
	//	fmt.Println(files)
	//}
	//
	//if fheader != nil {
	//	fmt.Println(fheader)
	//}
	//
	//file, err := GetFile("file", ht)
	//if err != nil {
	//	log.Println(err.Error())
	//	return nil, nil
	//}

	//fmt.Println("File: ", file.Name)


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

		err = file.Save("/tmp/NotasFiscais.xlsx")
		if err != nil {
			fmt.Println(err)
			return WriteStruct(NewUploadError(), http.StatusBadRequest)
		}

		dat, err := ioutil.ReadFile("/tmp/NotasFiscais.xlsx")
		if err != nil {
			fmt.Println("Ops!! :" + err.Error())
			return nil, err
		}

		filename := GenerateUniqueFilename("nota", "file.xlsx")
		err = Upload("us-east-1", "export-diniz", "notas/" + filename, filename,  bytes.NewReader(dat))
		if err != nil {
			fmt.Println(err)
			return WriteStruct(NewUploadS3Error(), http.StatusBadRequest)
		}

		urlFile, err := GetFileLink("us-east-1", "export-diniz", "notas/" + filename)
		if err != nil {
			fmt.Println(err)
			return WriteStruct(NewUploadS3Error(), http.StatusBadRequest)
		}
		hds := make(map[string]string)
		hds["Access-Control-Allow-Origin"] = "*"
		hds["Access-Control-Allow-Headers"] = "*"
		hds["Access-Control-Allow-Methods"] = "*"
		hds["Access-Control-Allow-Credentials"] = "true"
		return WriteStructWithHeader(APIResponseUrl{Success: true, Url:urlFile},http.StatusOK,hds)

		//text := ""
		//for _, row := range sheet.Rows {
		//	for _, cell := range row.Cells {
		//		text += cell.String()+";"
		//	}
		//	text += "\r\n"
		//}
		//
		//bytes := []byte(text)
		//err = ioutil.WriteFile("/tmp/teste.csv", bytes, 0777)
		//// handle this error
		//if err != nil {
		//	// print it out
		//	fmt.Println(err)
		//}
		//bytes, err := ioutil.ReadAll(fileXls)
		//if err != nil {
		//	fmt.Println("Ops!! :" + err.Error())
		//	return nil, err
		//}

		//hds := make(map[string]string)
		//hds["Access-Control-Allow-Origin"] = "*"
		//hds["Access-Control-Allow-Headers"] = "*"
		//hds["Access-Control-Allow-Methods"] = "*"
		//hds["Access-Control-Allow-Credentials"] = "true"
		//hds["Content-type"] = "application/csv"
		//hds["Content-Description"] = "File Transfer"
		//hds["Content-Disposition"] = "attachment; filename=teste.csv"
		//hds["Content-Length"] = strconv.Itoa(len(string(bytes)))
		//hds["filename"] = "teste.csv"
		//
		//fmt.Println("retornou o donwload o arquivo ",  http.StatusAccepted)
		//return WriteBinary(bytes, hds, http.StatusAccepted)
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
	runTeste("nota11.xml")
	//lambda.Start(run)
	//db, err := createSchema()
	//if err != nil {
	//	fmt.Println("Falha ao abrir conexão.")
	//}
	//
	//downloadNota(db)


}



type APIResponse struct {
	Success bool `json:"success"`
}

type APIResponseUrl struct {
	Success bool `json:"success"`
	Url string `json:"url"`
}

type APIRequest struct {
	file     []byte
	fileName string
	password string
}


type File struct{
	Name string
	Content []byte
	Header textproto.MIMEHeader
}

type Error struct {
	StatusCode int  `json:"status_code"`
	Success bool `json:"success"`
	ErrorType    string `json:"error"`
	Message string `json:"message"`
	Errors map[string][]string `json:"errors,omitempty"`
}

func NewDbError()*Error {
	er := Error{}
	er.Success = false
	er.ErrorType = "db_error"
	er.Message = "Erro connection db"
	er.StatusCode = http.StatusBadRequest
	return &er
}

func NewFileOpenError()*Error {
	er := Error{}
	er.Success = false
	er.ErrorType = "file_open_error"
	er.Message = "Erro file open"
	er.StatusCode = http.StatusBadRequest
	return &er
}

func NewFileSaveError()*Error {
	er := Error{}
	er.Success = false
	er.ErrorType = "file_save_error"
	er.Message = "Erro file save"
	er.StatusCode = http.StatusBadRequest
	return &er
}

func NewUploadS3Error()*Error {
	er := Error{}
	er.Success = false
	er.ErrorType = "upload_s3_error"
	er.Message = "Erro upload para s3"
	er.StatusCode = http.StatusBadRequest
	return &er
}

func NewUploadError()*Error {
	er := Error{}
	er.Success = false
	er.ErrorType = "upload_error"
	er.Message = "Erro upload"
	er.StatusCode = http.StatusBadRequest
	return &er
}

// WriteStructWithHeader - Generate APIGatewayProxyResponse to be return
func WriteStructWithHeader(data interface{}, statusCode int, headers map[string]string) (*events.APIGatewayProxyResponse, error) {

	bytes, err := json.Marshal(data)
	w := core.NewProxyResponseWriter()
	w.Header().Set("Content-Type", "application/json")

	for key, header := range headers {
		w.Header().Set(key, header)
	}

	w.WriteHeader(statusCode)
	w.Write(bytes)
	response, err := write(w)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func WriteStruct(data interface{}, statusCode int) (*events.APIGatewayProxyResponse, error) {

	bytes, err := json.Marshal(data)
	w := core.NewProxyResponseWriter()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(bytes)
	resp, err := write(w)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func WriteBinary(data []byte, headers map[string]string, statusCode int) (*events.APIGatewayProxyResponse, error) {

	w := core.NewProxyResponseWriter()
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(statusCode)
	w.Write(data)

	resp, err := write(w)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func WriteString(data string, statusCode int) (*events.APIGatewayProxyResponse, error) {

	w := core.NewProxyResponseWriter()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(data))

	resp, err := write(w)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func write(w *core.ProxyResponseWriter) (*events.APIGatewayProxyResponse, error) {

	resp, err := w.GetProxyResponse()
	if err != nil {
		return nil, err
	}
	return &resp, nil
}


// BasicAuth - Generate Base64 according with 'username' and 'password'.
func GetBasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " +base64.StdEncoding.EncodeToString([]byte(auth))
}

func ParseHttpRequest(r events.APIGatewayProxyRequest)(*http.Request,error){

	decodedBody := []byte(r.Body)

	params := url.Values{}
	for k, v := range r.QueryStringParameters {
		params.Add(k, v)
	}

	if r.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(r.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	req,err := http.NewRequest(r.HTTPMethod,r.Path+"?"+params.Encode(),bytes.NewReader(decodedBody))

	if err !=nil {
		return nil, err
	}

	req.Header = make(map[string][]string)

	for k, v := range r.Headers {
		if k == "content-type" || k == "Content-Type" {
			req.Header.Set(k, v)
		}
	}

	return req,nil
}


func GetFile(field string,req *http.Request)(*File,error){

	file, fh,err := req.FormFile(field)
	if err!=nil{
		return nil,err
	}
	buf := bytes.Buffer{}

	if _, err := io.Copy(&buf, file); err != nil {
		return nil, err
	}

	return &File{Name:fh.Filename,Content:buf.Bytes(),Header:fh.Header},nil
}

func Upload(region string, bucket string, key string, filename string, payload *bytes.Reader) error {

	//select Region to use.
	conf := aws.Config{Region: aws.String(region)}
	sess := session.New(&conf)
	svc := s3manager.NewUploader(sess)


	result, err := svc.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        payload,
		ContentType: aws.String(mime.TypeByExtension(filepath.Ext(filename))),
	})
	if err != nil {
		return err
	}

	log.Printf("Successfully uploaded %s to %s\n", filename, result.Location)

	return nil
}

func GenerateUniqueFilename(prefix, identifier string) (fileName string){
	time := strconv.Itoa(int(time.Now().UnixNano()))
	fileName = prefix + "_" + time + "_" + identifier
	return fileName
}


func GetFileLink(region string, bucket string, key string) (string, error) {

	log.Println("Region: ", region)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	svc := s3.New(sess)

	log.Println("Bucket: ", bucket)
	log.Println("Key: ", key)
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	req, _ := svc.GetObjectRequest(params)

	_, err = svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Printf(err.Error())
		return "", nil
	}

	urlPathS3, err := req.Presign(60 * time.Minute) // Set link expiration time
	if err != nil {
		return "", err
	}

	return url.QueryEscape(urlPathS3), err
}
