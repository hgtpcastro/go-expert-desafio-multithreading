package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type (
	DadosCep map[string]any

	RespostaApi struct {
		Api      string
		Conteudo DadosCep
	}
)

func (r RespostaApi) String() string {
	return fmt.Sprintf("Resposta da Api: %s, com o seguinte conteudo: %+v\n", r.Api, r.Conteudo)
}

func main() {
	cep := "01153000"
	listaApi := []string{
		"https://brasilapi.com.br/api/cep/v1/" + cep,
		"http://viacep.com.br/ws/" + cep + "/json/",
	}

	// Cria um canal para os resultados
	canalRespostaApi := make(chan RespostaApi)

	// Cria um WaitGroup para esperar todas as goroutines terminarem
	wg := sync.WaitGroup{}

	// Define o número de goroutines de acordo com a lista de API's
	wg.Add(len(listaApi))

	// Inicia as goroutines
	for _, urlApi := range listaApi {
		go obterDadosApiCep(urlApi, &wg, canalRespostaApi)
	}

	// Cria uma goroutine anônima para fechar o canal quando todas as outras goroutines terminarem
	go func() {
		wg.Wait()
		close(canalRespostaApi)
	}()

	// Define um temporizador para 1 segundo
	tempoEspera := time.After(1 * time.Second)

	// Agora, use select para aguardar o resultado ou o timeout
	select {
	case respostaApi := <-canalRespostaApi:
		fmt.Println(respostaApi.String())
	case <-tempoEspera:
		fmt.Println("Erro: Timeout atingido")
	}

}

func obterDadosApiCep(url string, wg *sync.WaitGroup, canalResultado chan RespostaApi) {
	defer wg.Done()

	// Simula timeout
	// time.Sleep(1 * time.Second)

	resposta, erro := http.Get(url)
	if erro != nil {
		// Envia o resultado para o canal
		conteudo := map[string]any{"erro": erro.Error()}
		canalResultado <- RespostaApi{
			Api:      url,
			Conteudo: conteudo,
		}
	}
	defer resposta.Body.Close()

	conteudoResposta, erro := io.ReadAll(resposta.Body)
	if erro != nil {
		// Envia o resultado para o canal
		conteudo := map[string]any{"erro": erro.Error()}
		canalResultado <- RespostaApi{
			Api:      url,
			Conteudo: conteudo,
		}
	}

	var dadosCep DadosCep
	if erro := json.Unmarshal(conteudoResposta, &dadosCep); erro != nil {
		// Envia o resultado para o canal
		conteudo := map[string]any{"erro": erro.Error()}
		canalResultado <- RespostaApi{
			Api:      url,
			Conteudo: conteudo,
		}
	}

	// Envia o resultado para o canal
	canalResultado <- RespostaApi{
		Api:      url,
		Conteudo: dadosCep,
	}

}
