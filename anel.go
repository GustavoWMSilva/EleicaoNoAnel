// LORENZO MORE
// GUSTAVO SILVA
// Código exemplo para o trabalho de sistemas distribuídos (eleição em anel)
// By Cesar De Rose - 2022

package main

import (
    "fmt"
    "sync"
)

type mensagem struct {
    tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmação da eleição)
    corpo [4]int // conteúdo da mensagem para colocar os IDs (usar um tamanho compatível com o número de processos no anel)
}

var (
    chans = []chan mensagem{ // vetor de canais para formar o anel de eleição - chans[0], chans[1] e chans[2] ...
        make(chan mensagem),
        make(chan mensagem),
        make(chan mensagem),
        make(chan mensagem),
    }
    controle = make(chan int)
    wg       sync.WaitGroup // wg é usado para esperar o programa terminar
)

func ElectionControler(in chan int) {
    defer wg.Done()

    var temp mensagem

    // comandos para o anel iniciam aqui

    // mudar o processo 0 - canal de entrada 3 - para falho (definir mensagem tipo 2 para isso)

    temp.tipo = 2
    chans[3] <- temp
    fmt.Printf("Controle: mudar o processo 0 para falho\n")

    fmt.Printf("Controle: confirmação de quem falhou: %d\n", <-in) // receber e imprimir confirmação
    fmt.Println("\n   Processo controlador concluído\n")

    temp.tipo = 3
    chans[0] <- temp
    fmt.Printf("Controle: processo 1, convoque nova eleicao\n")

    fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação
    fmt.Println("\n   Processo controlador concluído\n")

    temp.tipo = 2
    chans[2] <- temp
    fmt.Printf("Controle: mudar o processo 3 para falho\n")

    fmt.Printf("Controle: confirmação de quem falhou: %d\n", <-in) // receber e imprimir confirmação
    fmt.Println("\n   Processo controlador concluído\n")

    temp.tipo = 3
    chans[0] <- temp
    fmt.Printf("Controle: processo 1, convoque nova eleicao\n")

    fmt.Printf("Controle: lider atual %d\n", <-in) // receber e imprimir confirmação
    fmt.Println("\n   Processo controlador concluído\n")

   
    ///Ativando novamente

    temp.tipo = 3
    chans[3] <- temp
    fmt.Printf("Controle: ativar o processo 0\n")

    fmt.Printf("Controle: lider atual %d\n", <-in) // receber e imprimir confirmação
    fmt.Println("\n   Processo controlador concluído\n")

   
    temp.tipo = 3
    chans[2] <- temp
    fmt.Printf("Controle: ativar o processo 3\n")

    fmt.Printf("Controle: lider atual %d\n", <-in) // receber e imprimir confirmação
    fmt.Println("\n   Processo controlador concluído\n")

}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
    defer wg.Done()

    // variáveis locais que indicam se este processo é o líder e se está ativo

    var actualLeader int
    var bFailed bool = false // todos iniciam sem falha

    actualLeader = leader // indicação do líder que foi passada como parâmetro

    for i := 1; i < 20; i++ {
        temp := <-in // ler mensagem
        if(temp.tipo < 5 ){
            fmt.Printf("Processo %2d: recebi mensagem %d, [ %d, %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])
        }
        switch temp.tipo {
        // torna falho
        case 2:
            {
                bFailed = true
                fmt.Printf("Processo %2d: falhei, enviando mensagem pro controle\n", TaskId)
                fmt.Printf("Processo %2d: líder atual %d\n", TaskId, leader)
                leader = -5
                controle <- TaskId
            }
        // volta como era antes
        case 3:
            {
                // processo ativo
                bFailed = false
                leader = -5
                temp.tipo = 4
                for i := range temp.corpo {
                    temp.corpo[i] = -5 // desativa todos
                }
                temp.corpo[TaskId] = TaskId
                fmt.Printf("Processo %2d: colocando meu id na mensagem: [ %d, %d, %d, %d ]\n", TaskId, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])
                out <- temp
            }
        // colocar ID no corpo
        case 4:
            {
                if temp.corpo[TaskId] == TaskId {
                    // Percorra o array a partir do segundo elemento (índice 1) e compare cada elemento com o valor máximo
                    fmt.Printf("Processo %2d: recebi todas ids: [ %d, %d, %d, %d ]\n", TaskId, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])                  
                    for i := 0; i < len(temp.corpo); i++ {
                        if temp.corpo[i] > temp.corpo[0] {
                            temp.corpo[0] = temp.corpo[i]
                        }
                    }
                    leader = temp.corpo[0]
                    temp.corpo[1] = TaskId
                    temp.tipo = 5
                    fmt.Printf("Processo %2d: vencedor eleicao: processo %d \n", TaskId, leader)
                    fmt.Printf("Processo %2d: informando novo lider aos outros processos\n", TaskId)
                    fmt.Printf("Processo %2d: líder atualizado %d\n", TaskId, leader)
                } else if bFailed == false {
                    temp.corpo[TaskId] = TaskId
                } else{
                    fmt.Printf("Processo %2d:  Processo falho, pula\n", TaskId)
                }
                out <- temp
            }
        // mandando ID do novo líder para todos
        case 5:
            {
                if TaskId != temp.corpo[1] {
                    if bFailed == false {
                        leader = temp.corpo[0]
                        fmt.Printf("Processo %2d: Lider atualizado: %d \n", TaskId, leader)
                    }else {
                        fmt.Printf("Processo %2d:  Processo falho, pula\n", TaskId)
                    }
                } else {
                    temp.tipo = 6
                    controle <- leader
                }
                out <- temp

            }
        case 6:
            {
                break
            }
        default:
            {
                fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
                fmt.Printf("%2d: líder atual %d\n", TaskId, actualLeader)
            }
        }
    }
    fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {

    wg.Add(5) // Adicione uma contagem de quatro, um para cada goroutine

    // criar os processos do anel de eleição

    // func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
    go ElectionStage(0, chans[3], chans[0], 0) // este é o líder
    go ElectionStage(1, chans[0], chans[1], 0) // não é líder, é o processo 0
    go ElectionStage(2, chans[1], chans[2], 0) // não é líder, é o processo 0
    go ElectionStage(3, chans[2], chans[3], 0) // não é líder, é o processo 0

    fmt.Println("\n   Anel de processos criado")

    // criar o processo controlador

    go ElectionControler(controle)

    fmt.Println("\n   Processo controlador criado\n")

    wg.Wait() // Esperar pelas goroutines terminarem
}
