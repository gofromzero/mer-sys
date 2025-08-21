package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// 扩展已有的ProductStatus枚举，添加DELETED状态
const (
	ProductStatusDeleted ProductStatus = "deleted" // 已删除
)

// CategoryStatus 分类状态枚举
type CategoryStatus int

const (
	CategoryStatusActive   CategoryStatus = 1 // 启用
	CategoryStatusInactive CategoryStatus = 0 // 禁用
)

// ChangeOperation 变更操作类型
type ChangeOperation string

const (
	ChangeOperationCreate      ChangeOperation = "create"
	ChangeOperationUpdate      ChangeOperation = "update"
	ChangeOperationDelete      ChangeOperation = "delete"
	ChangeOperationStatusChange ChangeOperation = "status_change"
)

// ProductImage 商品图片
type ProductImage struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	AltText   string `json:"alt_text,omitempty"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

// ProductImages 商品图片数组类型，实现数据库序列化
type ProductImages []ProductImage

// Value 实现 driver.Valuer 接口
func (p ProductImages) Value() (driver.Value, error) {
	if len(p) == 0 {
		return json.Marshal([]ProductImage{})
	}
	return json.Marshal(p)
}

// Scan 实现 sql.Scanner 接口
func (p *ProductImages) Scan(value interface{}) error {
	if value == nil {
		*p = ProductImages{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	}
	return nil
}

// StringArray 字符串数组类型，用于商品标签
type StringArray []string

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	}
	return nil
}

// EnhancedProduct 增强的商品结构，扩展原有Product
type EnhancedProduct struct {
	ID           uint64         `json:"id" db:"id"`
	TenantID     uint64         `json:"tenant_id" db:"tenant_id"`
	MerchantID   uint64         `json:"merchant_id" db:"merchant_id"`
	Name         string         `json:"name" db:"name"`
	Description  string         `json:"description" db:"description"`
	
	// 扩展字段
	CategoryID   *uint64       `json:"category_id,omitempty" db:"category_id"`
	CategoryPath string        `json:"category_path,omitempty" db:"category_path"`
	Tags         StringArray   `json:"tags" db:"tags"`
	
	// 价格信息 - 使用JSON格式存储
	Price        json.RawMessage `json:"price" db:"price"` // 存储Money JSON
	RightsCost   int64          `json:"rights_cost" db:"rights_cost"` // 以分为单位
	
	// 库存信息 - 继续使用原有格式
	InventoryInfo *InventoryInfo `json:"inventory_info" db:"inventory_info"`
	
	// 状态和图片
	Status       ProductStatus  `json:"status" db:"status"`
	Images       ProductImages  `json:"images" db:"images"`
	Version      int           `json:"version" db:"version"`
	
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}

// GetPriceMoney 获取价格Money结构
func (ep *EnhancedProduct) GetPriceMoney() (*Money, error) {
	var money Money
	err := json.Unmarshal(ep.Price, &money)
	if err != nil {
		return nil, err
	}
	return &money, nil
}

// SetPriceMoney 设置价格Money结构
func (ep *EnhancedProduct) SetPriceMoney(money *Money) error {
	priceJSON, err := json.Marshal(money)
	if err != nil {
		return err
	}
	ep.Price = priceJSON
	return nil
}

// ToLegacyProduct 转换为原有Product结构
func (ep *EnhancedProduct) ToLegacyProduct() *Product {
	money, _ := ep.GetPriceMoney()
	var priceAmount float64
	var priceCurrency string
	
	if money != nil {
		priceAmount = float64(money.Amount) / 100 // 转换为元
		priceCurrency = money.Currency
	}
	
	return &Product{
		ID:            ep.ID,
		TenantID:      ep.TenantID,
		MerchantID:    ep.MerchantID,
		Name:          ep.Name,
		Description:   ep.Description,
		PriceAmount:   priceAmount,
		PriceCurrency: priceCurrency,
		RightsCost:    float64(ep.RightsCost) / 100, // 转换为元
		InventoryInfo: ep.InventoryInfo,
		Status:        ep.Status,
		CreatedAt:     ep.CreatedAt,
		UpdatedAt:     ep.UpdatedAt,
	}
}

// ProductCategory 商品分类实体
type ProductCategory struct {
	ID        uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID  uint64         `json:"tenant_id" gorm:"not null;index:idx_tenant_parent"`
	Name      string         `json:"name" gorm:"not null;size:100"`
	ParentID  *uint64        `json:"parent_id,omitempty" gorm:"index:idx_tenant_parent"`
	Level     int            `json:"level" gorm:"not null;default:1;index:idx_level"`
	Path      string         `json:"path" gorm:"not null;size:500;index:idx_path"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	Status    CategoryStatus `json:"status" gorm:"not null;default:1"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (ProductCategory) TableName() string {
	return "product_categories"
}

// ProductHistory 商品变更历史实体
type ProductHistory struct {
	ID        uint64          `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID  uint64          `json:"tenant_id" gorm:"not null;index:idx_tenant_product"`
	ProductID uint64          `json:"product_id" gorm:"not null;index:idx_tenant_product"`
	Version   int             `json:"version" gorm:"not null;index:idx_version"`
	FieldName string          `json:"field_name" gorm:"not null;size:100"`
	OldValue  string          `json:"old_value,omitempty" gorm:"type:text"`
	NewValue  string          `json:"new_value,omitempty" gorm:"type:text"`
	Operation ChangeOperation `json:"operation" gorm:"not null;size:20"`
	ChangedBy uint64          `json:"changed_by" gorm:"not null"`
	ChangedAt time.Time       `json:"changed_at" gorm:"autoCreateTime;index:idx_changed_at"`
}

// TableName 指定表名
func (ProductHistory) TableName() string {
	return "product_histories"
}

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	Name         string        `json:"name" validate:"required,max=255"`
	Description  string        `json:"description,omitempty"`
	CategoryID   *uint64       `json:"category_id,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
	Price        Money         `json:"price" validate:"required"`
	RightsCost   int64         `json:"rights_cost" validate:"min=0"`
	Inventory    InventoryInfo `json:"inventory" validate:"required"`
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	Name        string        `json:"name,omitempty" validate:"max=255"`
	Description string        `json:"description,omitempty"`
	CategoryID  *uint64       `json:"category_id,omitempty"`
	Tags        []string      `json:"tags,omitempty"`
	Price       *Money        `json:"price,omitempty"`
	RightsCost  *int64        `json:"rights_cost,omitempty" validate:"min=0"`
	Inventory   *InventoryInfo `json:"inventory,omitempty"`
}

// UpdateProductStatusRequest 更新商品状态请求
type UpdateProductStatusRequest struct {
	Status ProductStatus `json:"status" validate:"required,oneof=draft active inactive deleted"`
}

// ProductListRequest 商品列表查询请求
type ProductListRequest struct {
	Page       int           `json:"page" validate:"min=1"`
	PageSize   int           `json:"page_size" validate:"min=1,max=100"`
	CategoryID *uint64       `json:"category_id,omitempty"`
	Status     ProductStatus `json:"status,omitempty"`
	Keyword    string        `json:"keyword,omitempty"`
	SortBy     string        `json:"sort_by,omitempty" validate:"oneof=created_at updated_at name price"`
	SortOrder  string        `json:"sort_order,omitempty" validate:"oneof=asc desc"`
}

// ProductBatchOperationRequest 批量操作请求
type ProductBatchOperationRequest struct {
	ProductIDs []uint64      `json:"product_ids" validate:"required,min=1"`
	Operation  string        `json:"operation" validate:"required,oneof=activate deactivate delete"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name      string  `json:"name" validate:"required,max=100"`
	ParentID  *uint64 `json:"parent_id,omitempty"`
	SortOrder int     `json:"sort_order,omitempty"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name      string  `json:"name,omitempty" validate:"max=100"`
	ParentID  *uint64 `json:"parent_id,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
	Status    *CategoryStatus `json:"status,omitempty"`
}

// UploadImageRequest 上传图片请求
type UploadImageRequest struct {
	AltText   string `json:"alt_text,omitempty"`
	SortOrder int    `json:"sort_order,omitempty"`
	IsPrimary bool   `json:"is_primary,omitempty"`
}

// ProductResponse 商品响应
type ProductResponse struct {
	Product
	Category *ProductCategory `json:"category,omitempty"`
}

// ProductListResponse 商品列表响应
type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// CategoryTreeResponse 分类树响应
type CategoryTreeResponse struct {
	ProductCategory
	Children []CategoryTreeResponse `json:"children,omitempty"`
}