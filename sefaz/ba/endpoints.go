package ba

const (
	baseURL = "https://nfe.sefaz.ba.gov.br"

	pathAccessKey = "/servicos/nfce/Modulos/Geral/NFCEC_consulta_chave_acesso.aspx"
	pathCaptcha   = "/servicos/nfce/Modulos/AntiRobo/NFCEC_anti_robo.aspx"
	pathDANFE     = "/servicos/nfce/Modulos/Geral/NFCEC_consulta_danfe.aspx"
	pathPrint     = "/servicos/nfce/Modulos/Geral/Frm_Imprimir_parcial.aspx"
)

const (
	fieldAccessKey = "txt_chave_acesso"
	fieldCaptcha   = "txt_cod_antirobo"
	fieldSubmit    = "btn_consulta_completa"
	fieldViewTabs  = "btn_visualizar_abas"
)

const (
	maxCaptchaRetries = 3
	requestsPerSecond = 2
)
