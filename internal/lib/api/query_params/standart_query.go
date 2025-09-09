package query_params

import (
	"fmt"
	"github.com/go-playground/validator"
	"log/slog"
	"net/url"
	"strconv"
)

// BaseQueryParams содержит общие параметры для всех запросов
type BaseQueryParams struct {
	Search string `validate:"omitempty"`     // Поисковая строка, необязательная
	Limit  int    `validate:"gte=1,lte=100"` // Лимит записей (1–100)
	Offset int    `validate:"gte=0"`         // Смещение для пагинации
	Page   int    `validate:"gte=1"`         // Номер страницы
}

// SortParam представляет один параметр сортировки
type SortParam struct {
	Field string // Поле сортировки (например, "id")
	Order string // Порядок сортировки ("asc" или "desc")
}

// ListQueryParams объединяет базовые параметры и параметры сортировки
type ListQueryParams struct {
	BaseQueryParams
	SortParams []SortParam // Список параметров сортировки
}

// QueryParamsParser — интерфейс для кастомной обработки сортировки
type QueryParamsParser interface {
	ParseSortParams(query url.Values, log *slog.Logger) ([]SortParam, error)
}

// DefaultSortParser реализует парсинг сортировки
type DefaultSortParser struct {
	ValidSortFields []string // Список разрешённых полей для сортировки
}

// ParseSortParams ищет в query параметры, совпадающие с ValidSortFields,
// и интерпретирует их значения как порядок сортировки (asc/desc)
func (p *DefaultSortParser) ParseSortParams(query url.Values, log *slog.Logger) ([]SortParam, error) {
	var sortParams []SortParam

	// Проходим по всем разрешённым полям сортировки
	for _, field := range p.ValidSortFields {
		// Проверяем, есть ли поле в query-параметрах
		order := query.Get(field)
		if order == "" {
			continue // Поле не указано в query, пропускаем
		}

		// Проверяем, что порядок сортировки валидный
		if order != "asc" && order != "desc" {
			log.Warn("Invalid sort order", "field", field, "order", order)
			return nil, fmt.Errorf("invalid sort order for field %s: %s", field, order)
		}

		// Добавляем параметр сортировки
		sortParams = append(sortParams, SortParam{Field: field, Order: order})
	}

	return sortParams, nil
}

// ParseStandardQueryParams парсит стандартные параметры и делегирует сортировку парсеру
func ParseStandardQueryParams(query url.Values, log *slog.Logger, parser QueryParamsParser) (ListQueryParams, error) {
	log = log.With(
		slog.String("query", query.Encode()),
		slog.String("operation", "ParseStandardQueryParams"))

	// Инициализируем структуру с значениями по умолчанию
	params := ListQueryParams{
		BaseQueryParams: BaseQueryParams{
			Limit:  10, // Дефолтный лимит
			Offset: 0,  // Дефолтное смещение
			Page:   1,  // Дефолтная страница
		},
	}

	// Парсим параметр search
	params.Search = query.Get("search")

	// Парсим параметр limit
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			log.Warn("Invalid limit", "error", err)
			return params, fmt.Errorf("invalid limit parameter: %w", err)
		}
		params.Limit = limit
	}

	// Парсим параметр page или offset
	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.Warn("Invalid page", "error", err)
			return params, fmt.Errorf("invalid page format: %w", err)
		}
		params.Page = page
		params.Offset = (page - 1) * params.Limit // Вычисляем offset из page
	} else if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			log.Warn("Invalid offset", "error", err)
			return params, fmt.Errorf("invalid offset format: %w", err)
		}
		params.Offset = offset
	}

	// Валидируем базовые параметры
	if err := validator.New().Struct(params.BaseQueryParams); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		log.Error("Error validating query params", "err", validationErrors)
		return params, fmt.Errorf("error validating query params: %w", err)
	}

	// Парсим сортировку, если передан парсер
	if parser != nil {
		sortParams, err := parser.ParseSortParams(query, log)
		if err != nil {
			return params, err
		}
		params.SortParams = sortParams
	}

	return params, nil
}
