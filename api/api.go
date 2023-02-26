package api

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/external-fun/grpc-server/proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
)

type DatabaseExporterService struct {
	proto.UnimplementedDatabaseExporterServer
	db *sql.DB
}

func insertRow(db *sql.DB, table string, field string, value string) (int, error) {
	id := 0
	err := db.QueryRow(fmt.Sprintf(`INSERT INTO public."%s"(%s) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id`, table, field), value).Scan(&id)
	if err == nil {
		return id, nil
	} else if err == sql.ErrNoRows {
		err := db.QueryRow(fmt.Sprintf(`SELECT id FROM public."%s" WHERE %s = $1`, table, field), value).Scan(&id)
		return id, err
	} else {
		return 0, err
	}
}

func insertBrand(db *sql.DB, row *proto.Row) (int, error) {
	return insertRow(db, "Brand", "name", row.BrandName)
}

func insertCategory(db *sql.DB, row *proto.Row) (int, error) {
	return insertRow(db, "Category", "name", row.CategoryName)
}

func insertSize(db *sql.DB, row *proto.Row) (int, error) {
	return insertRow(db, "Size", "name", row.SizeName)
}

func InsertRow(db *sql.DB, row *proto.Row) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	brandId, err := insertBrand(db, row)
	if err != nil {
		return err
	}

	categoryId, err := insertCategory(db, row)
	if err != nil {
		return err
	}

	sizeId, err := insertSize(db, row)
	if err != nil {
		return err
	}

	// TODO: have duplicates?
	_, err = db.Exec(`INSERT INTO public."Clothes"(id, name, brand_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, row.ClothesId, row.ClothesName, brandId)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO public."Record"(quantity, size_id, clothes_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, row.Quantity, sizeId, row.ClothesId)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO public."ClothesAndCategory"(clothes_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, row.ClothesId, categoryId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (service *DatabaseExporterService) UploadRows(stream proto.DatabaseExporter_UploadRowsServer) error {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		row, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			fmt.Println(err)
			continue
		}

		err = InsertRow(service.db, row)
		if err != nil {
			log.Println(err)
		}
	}
}

func NewDatabaseExporterService(db *sql.DB) *DatabaseExporterService {
	return &DatabaseExporterService{
		db: db,
	}
}

func ListenAndServe(address string, service *DatabaseExporterService) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("here")
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterDatabaseExporterServer(grpcServer, service)
	return grpcServer.Serve(listener)
}
