pipeline {
    agent any

    environment {
        // Aponta os endereços locais
        ENVOY_URL      = 'localhost:10000'
        POSTGRES_DB    = 'postgresql://agente_go:senha_segura@localhost:5432/cambio_vector'
        REDIS_URL      = 'redis://localhost:6379'
        
        // FORÇA O JENKINS A ENXERGAR OS COMPILADORES DA SUA MÁQUINA
        PATH           = "/usr/local/bin:/usr/bin:/bin:/usr/local/go/bin:/home/devildev/.cargo/bin:$PATH"
        
        RUST_BACKTRACE = '1'
        GOENV          = 'production'
    }

    stages {
        stage('1. Validação de Contrato Semântico') {
            steps {
                echo '=== ESTÁGIO 1: Checando integridade do arquivo .proto ==='
                sh 'buf lint contracts/ --path contracts/cambio/regulatorio/v1/cambio.proto'
            }
        }

        stage('2. Compilação Bare-Metal (Rust & Go)') {
            steps {
                echo '=== ESTÁGIO 2: Compilando motores nativos para o Backplane ==='
                sh 'cd svc-stream-rust && cargo +stable build --release'
                sh 'cd svc-agente-go && go build -o agente_roteador main.go'
            }
        }

        stage('3. Auditoria do Espaço Vetorial (Postgres 16)') {
            steps {
                echo '=== ESTÁGIO 3: Validando integridade do PgVector e Índices HNSW ==='
                // Alterado de 'pgvector' para 'vector' para refletir a compilação nativa
                sh 'psql postgresql://agente_go:senha_segura@localhost:5432/cambio_vector -c "SELECT extname, extversion FROM pg_extension WHERE extname = \'vector\';"'
            }
        }

        stage('4. Simulação de Stream Regulatório e Prova do Cosseno') {
            steps {
                echo '=== ESTÁGIO 4: Injetando transações de Dropshipping B2B via Stream ==='
                sh 'python3 analytics-python/simular_estresse.py --target ${ENVOY_URL}'
            }
        }
    }

    post {
        success {
            echo '=== CANTEIRO APROVADO: A arquitetura está estável e o cálculo do cosseno é válido. Liberando Merge no GitHub. ==='
        }
        failure {
            echo '=== CANTEIRO BLOQUEADO: Falha crítica detectada no fluxo local. Verifique os logs de latência ou quebra de contrato. ==='
        }
    }
}

