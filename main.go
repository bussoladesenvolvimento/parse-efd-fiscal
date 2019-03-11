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
	"io"
	"io/ioutil"
	"mime"
	"net/textproto"
	"net/url"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
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



func runTeste(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error){

	mydata := []byte("all my data I want to write to a file")
	err := ioutil.WriteFile("/tmp/teste.xlsx", mydata, 0777)
	// handle this error
	if err != nil {
		// print it out
		fmt.Println(err)
	}

	//
	//file, err := os.Create("/tmp/teste.txt")
	//if err != nil {
	//	fmt.Println("Cannot create file")
	//}
	//defer file.Close()
	//
	//bytes, _ := ioutil.ReadAll(file)

	hds := make(map[string]string)
	hds["Access-Control-Allow-Origin"] = "*"
	hds["Access-Control-Allow-Headers"] = "*"
	hds["Access-Control-Allow-Methods"] = "*"
	hds["Access-Control-Allow-Credentials"] = "true"
	hds["Content-type"] = "application/xlsx"
	hds["Content-Description"] = "File Transfer"
	hds["Content-Disposition"] = "attachment; filename=teste.xlsx"
	hds["Content-Length"] = strconv.Itoa(len(string(mydata)))
	hds["filename"] = "teste.xlsx"

	return WriteBinary(mydata, hds, http.StatusOK)
}

func run(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error){

	ht, err := ParseHttpRequest(r)

	file, err := GetFile("file", ht)
	if err != nil {
		log.Println(err.Error())
		return nil, nil
	}

	db, err := createSchema()
	if err != nil {
		fmt.Println("Falha ao abrir conexão.")
		return WriteStruct(NewDbError(), http.StatusBadRequest)
	}

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
		ext := filepath.Ext(file.Name)

		if ext == ".xml" || ext == ".XML" {

			SpedRead.InsertXml(&notas, file.Content, dialect, conexao, digitos)

			//go InsertXml(file, dialect, conexao, digitosCodigo)
		}

		//SpedRead.RecursiveSpeds("./speds", &notas, dialect, conexao, digitos)
		// Pega cada arquivo e ler linha a linha e envia para o banco de dados
		fmt.Println("Final processamento ", time.Now())
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

		return WriteStruct(&APIResponseUrl{Success: true, Url:urlFile}, 200)

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

	lambda.Start(run)
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
