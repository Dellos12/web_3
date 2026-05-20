package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "net"

    _ "github.com/lib/pq"
    "google.golang.org/grpc"

    // Pacote local gerado pelo protoc
    pb "svc-agente-go/regulatorio/v1"
)

type AgenteGoServer struct {
    pb.UnimplementedRoteadorFiscalServiceServer
    db *sql.DB
}

func (s *AgenteGoServer) ValidarE_RotearTransacao(ctx context.Context, req *pb.RequestTransacao) (*pb.ResponseDirecionamento, error) {
    // Formata vetor para PgVector
    var vetorFormatado string
    for i, val := range req.OperacaoEmbedding {
        if i == 0 {
            vetorFormatado += fmt.Sprintf("{%f", val)
        } else {
            vetorFormatado += fmt.Sprintf(",%f", val)
        }
    }
    vetorFormatado += "}"

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
        return &pb.ResponseDirecionamento{ConformidadeAprovada: false, RailEscolhido: pb.ResponseDirecionamento_RAIL_REJEITADO}, nil
    }

    similaridadeCosseno := 1.0 - cossenoDistancia
    log.Printf("📊 Similaridade com a matriz regulatória: %f", similaridadeCosseno)

    if similaridadeCosseno < 0.85 {
        return &pb.ResponseDirecionamento{
            TransacaoId:          req.TransacaoId,
            ConformidadeAprovada: false,
            RailEscolhido:        pb.ResponseDirecionamento_RAIL_REJEITADO,
        }, nil
    }

    valorImposto := req.ValorOperacaoBrl * ((aliquotaCbs + aliquotaIbs) / 100.0)

    return &pb.ResponseDirecionamento{
        TransacaoId:           req.TransacaoId,
        ConformidadeAprovada:  true,
        AliquotaCbsFederal:    aliquotaCbs,
        AliquotaIbsEstadual:   aliquotaIbs,
        ValorImpostoRetidoBrl: valorImposto,
        RailEscolhido:         pb.ResponseDirecionamento_RAIL_DESCENTRALIZADO_WEB3,
        HashAuditoriaEstado:   req.DocumentoFiscalSha256,
    }, nil
}

func main() {
    dsn := "postgresql://postgres:postgres@localhost:5432/cambio_vector?sslmode=disable"
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

    log.Println("🔀 Agente Orquestrador Go rodando na porta :50052...")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Falha no servidor gRPC Go: %v", err)
    }
}

