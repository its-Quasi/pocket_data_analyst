package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/employees?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Find top 10 titles by average salary between 1986 and 1993
	queryTopTitles := `
		SELECT t.title
		FROM titles t
		JOIN salaries s ON t.emp_no = s.emp_no
		WHERE s.from_date >= '1986-01-01' AND s.to_date <= '1993-12-31'
		GROUP BY t.title
		ORDER BY AVG(s.salary) DESC
		LIMIT 10`

	rows, err := db.Query(queryTopTitles)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var topTitles []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			log.Fatal(err)
		}
		topTitles = append(topTitles, title)
	}

	years := []int{1986, 1987, 1988, 1989, 1990, 1991, 1992, 1993}
	xLabels := make([]string, 0)
	for _, y := range years {
		xLabels = append(xLabels, fmt.Sprintf("%d", y))
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Evolución Salarial de los 10 Cargos Mejor Pagados (1986-1993)",
			Subtitle: "Promedio anual por cargo",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    true,
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: types.Bool(true),
		}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Año"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Salario Promedio"}),
	)
	line.SetXAxis(xLabels)

	for _, title := range topTitles {
		var values []opts.LineData
		for _, year := range years {
			var avgSalary sql.NullFloat64
			queryYearly := `
				SELECT AVG(s.salary)
				FROM salaries s
				JOIN titles t ON s.emp_no = t.emp_no
				WHERE t.title = ? AND YEAR(s.from_date) = ?`

			err := db.QueryRow(queryYearly, title, year).Scan(&avgSalary)
			if err != nil {
				log.Fatal(err)
			}

			if avgSalary.Valid {
				values = append(values, opts.LineData{Value: avgSalary.Float64})
			} else {
				values = append(values, opts.LineData{Value: 0})
			}
		}
		line.AddSeries(title, values)
	}

	os.MkdirAll("./sandbox_area/charts", 0755)
	fileName := fmt.Sprintf("salary_evolution_%d.html", time.Now().Unix())
	filePath := filepath.Join("./sandbox_area/charts", fileName)
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}

	err = line.Render(f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(filePath)
}
