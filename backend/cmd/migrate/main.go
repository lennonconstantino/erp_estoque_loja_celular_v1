// Comando migrate aplica/reverte as migrations SQL contra o banco apontado por
// DATABASE_URL. É o runner usado no pre-deploy do Railway (`/app/migrate up`) e
// é intercambiável com a imagem migrate/migrate do docker-compose: usa os mesmos
// arquivos `migrations/NNNNNN_nome.up.sql`/`.down.sql` e a mesma tabela de
// controle `schema_migrations`.
//
// Uso:
//
//	migrate up               # aplica todas as migrations pendentes
//	migrate down             # reverte a última migration
//	migrate down <n>         # reverte n migrations
//	migrate version          # imprime a versão atual (e se está "dirty")
//	migrate force <version>  # marca a versão atual sem rodar (destrava dirty)
//
// Flags:
//
//	-path   diretório das migrations (default: ./migrations, ou MIGRATIONS_PATH)
//	-url    string de conexão        (default: $DATABASE_URL)
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // carrega .env em dev; em produção usa env real.

	path := flag.String("path", getenv("MIGRATIONS_PATH", "./migrations"), "diretório das migrations")
	url := flag.String("url", os.Getenv("DATABASE_URL"), "string de conexão (DATABASE_URL)")
	flag.Parse()

	if *url == "" {
		log.Fatal("migrate: DATABASE_URL não definida (use -url ou a variável de ambiente)")
	}

	// O driver registrado é "pgx5"; aceitamos URLs postgres:// reescrevendo o
	// esquema para que o golang-migrate selecione o driver pgx/v5.
	dbURL := normalizarURL(*url)

	m, err := migrate.New("file://"+*path, dbURL)
	if err != nil {
		log.Fatalf("migrate: falha ao inicializar: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("migrate: erro ao fechar source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("migrate: erro ao fechar conexão: %v", dbErr)
		}
	}()

	args := flag.Args()
	cmd := "up"
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "up":
		run(m.Up)
	case "down":
		if len(args) > 1 {
			n, err := strconv.Atoi(args[1])
			if err != nil || n <= 0 {
				log.Fatalf("migrate: argumento inválido para down: %q", args[1])
			}
			run(func() error { return m.Steps(-n) })
		} else {
			run(func() error { return m.Steps(-1) })
		}
	case "version":
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("nenhuma migration aplicada")
			return
		}
		if err != nil {
			log.Fatalf("migrate: %v", err)
		}
		fmt.Printf("versão=%d dirty=%t\n", v, dirty)
	case "force":
		if len(args) < 2 {
			log.Fatal("migrate: force exige uma versão (ex.: migrate force 9)")
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("migrate: versão inválida para force: %q", args[1])
		}
		if err := m.Force(v); err != nil {
			log.Fatalf("migrate: force falhou: %v", err)
		}
		fmt.Printf("versão forçada para %d\n", v)
	default:
		log.Fatalf("migrate: comando desconhecido %q (use up|down|version|force)", cmd)
	}
}

// run executa uma ação de migration tratando ErrNoChange como sucesso silencioso
// (idempotência: rodar `up` num banco já migrado não é erro).
func run(action func() error) {
	if err := action(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("migrate: nada a fazer (banco já está na versão mais recente)")
			return
		}
		log.Fatalf("migrate: %v", err)
	}
	fmt.Println("migrate: concluído com sucesso")
}

// normalizarURL garante o esquema pgx5:// esperado pelo driver registrado,
// aceitando as formas usuais postgres:// e postgresql://.
func normalizarURL(url string) string {
	switch {
	case len(url) > 11 && url[:11] == "postgres://":
		return "pgx5://" + url[11:]
	case len(url) > 13 && url[:13] == "postgresql://":
		return "pgx5://" + url[13:]
	default:
		return url
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
