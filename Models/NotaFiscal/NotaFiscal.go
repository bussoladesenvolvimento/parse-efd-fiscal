package NotaFiscal

import (
	"github.com/jinzhu/gorm"
	"time"
)

// Estrutura da nota fiscal eletronica
type NotaFiscal struct {
	gorm.Model
	NNF            string
	ChNFe          string
	NatOp          string
	IndPag         string
	Mod            string
	Serie          string
	DEmi           time.Time
	TpNF           string
	TpImp          string
	TpEmis         string
	CDV            string
	TpAmb          string
	FinNFe         string
	ProcEmi        string
	Emitente       Emitente
	EmitenteID     int
	Destinatario   Destinatario
	DestinatarioID int

	//ICMSTot
	ICMSTotVBC           float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVICMS         float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVICMSDeson    float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVBCST         float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVST           float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVProd         float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVFrete        float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVSeg          float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVDesc         float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVII           float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVIPI          float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVPIS          float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVOutro        float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVCOFINS       float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVNF           float64   `gorm:"type:decimal(19,3)"`
	ICMSTotVTotTrib      float64   `gorm:"type:decimal(19,3)"`

	Itens          []Item
}
