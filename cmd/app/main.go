package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type (
	DadosCep map[string]any

	RespostaApi struct {
		Api      string
		Conteudo DadosCep
	}

	CanalResposta chan RespostaApi
)

func (r RespostaApi) String() string {
	return fmt.Sprintf("Recebido da Api: %s, o seguinte conteudo: %+v", r.Api, r.Conteudo)
}

func main() {
	tempoResposta := 100 * time.Second
	cep := "01153000"

	canalRespostaApi1 := make(CanalResposta)
	canalRespostaApi2 := make(CanalResposta)

	ctx, cancelarReqisicao := context.WithTimeout(context.Background(), tempoResposta)
	defer cancelarReqisicao()

	go func() {
		dadosCep, erro := obterDadosCep(ctx, "https://brasilapi.com.br/api/cep/v1/"+cep)
		if erro != nil {
			panic(erro)
		}

		canalRespostaApi1 <- RespostaApi{
			Api:      "brasilapi",
			Conteudo: dadosCep,
		}

	}()

	go func() {
		dadosCep, erro := obterDadosCep(ctx, "http://viacep.com.br/ws/"+cep+"/json/")
		if erro != nil {
			panic(erro)
		}

		canalRespostaApi2 <- RespostaApi{
			Api:      "viacep",
			Conteudo: dadosCep,
		}
	}()

	for i := 0; i < 2; i++ {
		select {
		case respostaApi1 := <-canalRespostaApi1:
			fmt.Println(respostaApi1.String())
		case respostaApi2 := <-canalRespostaApi2:
			fmt.Println(respostaApi2.String())
		}
	}

}

func obterDadosCep(contexto context.Context, url string) (dadosCep DadosCep, erro error) {

	requisicao, erro := http.NewRequestWithContext(contexto, http.MethodGet, url, nil)
	if erro != nil {
		return nil, erro
	}

	resposta, erro := http.DefaultClient.Do(requisicao)
	if erro != nil {
		return nil, erro
	}
	defer resposta.Body.Close()

	conteudoResposta, erro := io.ReadAll(resposta.Body)
	if erro != nil {
		return nil, erro
	}

	json.Unmarshal(conteudoResposta, &dadosCep)

	return dadosCep, nil
}
