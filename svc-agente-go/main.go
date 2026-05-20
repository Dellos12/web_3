package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "net"

    _ "github.com/lib/pq" // Corrigido: driver Postgres
    "google.golang.org/grpc"
    
    pb "svc-agente-go/regulatorio/v1"
)

type AgenteGoServer struct {
    pb.UnimplementedRoteadorFiscalServiceServer
    db *sql.DB
}

// 📐 A PROVA DO COSSENO: Atualizado com os novos nomes do contrato do Buf
func (s *AgenteGoServer) ValidarERotearTransacao(ctx context.Context, req *pb.ValidarERotearTransacaoRequest) (*pb.ValidarERotearTransacaoResponse, error) {
    
    // Transforma o array de floats em uma string formatada para o tipo 'vector'
    var vetorFormatado string
    for i, val := range req.OperacaoEmbedding {
        if i == 0 {
            vetorFormatado += fmt.Sprintf("[%f", val)
        } else {
            vetorFormatado += fmt.Sprintf(",%f", val)
        }
    }
    vetorFormatado += "]"

    var cnaeCodigo string
    var aliquotaCbs, aliquotaIbs float64
    var cossenoDistancia float64

    query := `
        SELECT cnae_codigo, aliquota_cbs_base, aliquota_ibs_base, (regra_embedding <=> $1) as distancia
        FROM matriz_regulatoria 
        ORDER BY distancia ASC 
        LIMIT 1;`

    err := s.db.QueryRowContext(ctx, query, vetorFormatado).Scan(&cnaeCodigo, &aliquotaCbs, &aliquotaIbs, &cossenoDistancia)
    if err != nil {
        log.Printf("⚠️ Erro ao calcular cosseno no PgVector: %v", err)
        return &pb.ValidarERotearTransacaoResponse{
            ConformidadeAprovada: false,
            RailEscolhido: pb.ValidarERotearTransacaoResponse_RAIL_DESTINO_REJEITADO_UNSPECIFIED,
        }, nil
    }

    similaridadeCosseno := 1.0 - cossenoDistancia
    log.Printf("📊 Proximidade geométrica com a Reforma Tributária: %f", similaridadeCosseno)

    if similaridadeCosseno < 0.85 {
        return &pb.ValidarERotearTransacaoResponse{
            TransacaoId:          req.TransacaoId,
            ConformidadeAprovada: false,
            RailEscolhido:        pb.ValidarERotearTransacaoResponse_RAIL_DESTINO_REJEITADO_UNSPECIFIED,
        }, nil
    }

    valorImposto := req.ValorOperacaoBrl * ((aliquotaCbs + aliquotaIbs) / 100.0)

    return &pb.ValidarERotearTransacaoResponse{
        TransacaoId:           req.TransacaoId,
        ConformidadeAprovada:  true,
        AliquotaCbsFederal:    aliquotaCbs,
        AliquotaIbsEstadual:   aliquotaIbs,
        ValorImpostoRetidoBrl: valorImposto,
        RailEscolhido:         pb.ValidarERotearTransacaoResponse_RAIL_DESTINO_DESCENTRALIZADO_WEB3, 
        HashAuditoriaEstado:   req.DocumentoFiscalSha256,
    }, nil
}

func main() {
    dsn := "postgresql://agente_go:senha_segura@localhost:5432/cambio_vector?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("Falha ao abrir banco local: %v", err)
    }
    defer db.Close()

    lis, err := net.Listen("tcp", "127.0.0.1:50052")
    if err != nil {
        log.Fatalf("Falha ao escutar backplane em Go: %v", err)
    }

    grpcServer := grpc.NewServer()
    server := &AgenteGoServer{db: db}
    pb.RegisterRoteadorFiscalServiceServer(grpcServer, server)

    log.Println("🔀 Agente Orquestrador Go rodando localmente na porta :50052...")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Falha no servidor gRPC Go: %v", err)
    }
}

