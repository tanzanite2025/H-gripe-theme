package repository

import (
	"encoding/json"
	"strings"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	shippingdomain "commerce-platform/internal/domain/shipping"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const defaultShipmentWarrantyMonths = 12

type ShipmentRecordRepository struct {
	db *gorm.DB
}

type ShipmentRecordFilter struct {
	Keyword string
	Status  string
}

func NewShipmentRecordRepository(db *gorm.DB) *ShipmentRecordRepository {
	return &ShipmentRecordRepository{db: db}
}

func (r *ShipmentRecordRepository) WithTx(tx *gorm.DB) *ShipmentRecordRepository {
	return &ShipmentRecordRepository{db: tx}
}

func (r *ShipmentRecordRepository) FindByID(id uint) (*shippingdomain.ShipmentRecord, error) {
	return r.FindByOrderID(id)
}

func (r *ShipmentRecordRepository) FindByOrderID(orderID uint) (*shippingdomain.ShipmentRecord, error) {
	orderRecord, err := r.findShippedOrderByID(orderID, true)
	if err != nil {
		return nil, err
	}
	attachment, err := r.findAttachmentByOrderID(orderID)
	if err != nil && !IsRecordNotFound(err) {
		return nil, err
	}
	if IsRecordNotFound(err) {
		attachment = nil
	}
	return buildShipmentRecord(orderRecord, attachment), nil
}

func (r *ShipmentRecordRepository) FindByOrderNumber(orderNumber string) (*shippingdomain.ShipmentRecord, error) {
	orderRecord, err := r.findShippedOrderByNumber(orderNumber, 0)
	if err != nil {
		return nil, err
	}
	attachment, err := r.findAttachmentByOrderID(orderRecord.ID)
	if err != nil && !IsRecordNotFound(err) {
		return nil, err
	}
	if IsRecordNotFound(err) {
		attachment = nil
	}
	return buildShipmentRecord(orderRecord, attachment), nil
}

func (r *ShipmentRecordRepository) FindByOrderNumberForUser(orderNumber string, userID uint) (*shippingdomain.ShipmentRecord, error) {
	orderRecord, err := r.findShippedOrderByNumber(orderNumber, userID)
	if err != nil {
		return nil, err
	}
	attachment, err := r.findAttachmentByOrderID(orderRecord.ID)
	if err != nil && !IsRecordNotFound(err) {
		return nil, err
	}
	if IsRecordNotFound(err) {
		attachment = nil
	}
	return buildShipmentRecord(orderRecord, attachment), nil
}

// FindAll lists shipped orders and left-joins the optional after-sales
// attachment in application code. A shipping operation therefore never needs
// to create or update a row in shipment_records.
func (r *ShipmentRecordRepository) FindAll(page, pageSize int, filter ShipmentRecordFilter) ([]shippingdomain.ShipmentRecord, int64, error) {
	headers, err := r.findShippedOrderHeaders(filter.Keyword)
	if err != nil {
		return nil, 0, err
	}

	attachments, err := r.loadAttachments(orderIDs(headers))
	if err != nil {
		return nil, 0, err
	}

	filtered := make([]orderdomain.Order, 0, len(headers))
	for index := range headers {
		view := buildShipmentRecord(&headers[index], attachments[headers[index].ID])
		if shipmentRecordStatusMatches(view, filter.Status) {
			filtered = append(filtered, headers[index])
		}
	}

	total := int64(len(filtered))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return []shippingdomain.ShipmentRecord{}, total, nil
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	selected := filtered[start:end]
	fullOrders, err := r.loadOrdersWithItems(orderIDs(selected))
	if err != nil {
		return nil, 0, err
	}

	records := make([]shippingdomain.ShipmentRecord, 0, len(selected))
	for index := range selected {
		orderRecord := fullOrders[selected[index].ID]
		if orderRecord == nil {
			orderRecord = &selected[index]
		}
		records = append(records, *buildShipmentRecord(orderRecord, attachments[orderRecord.ID]))
	}
	return records, total, nil
}

func (r *ShipmentRecordRepository) GetStats() (map[string]interface{}, error) {
	headers, err := r.findShippedOrderHeaders("")
	if err != nil {
		return nil, err
	}
	attachments, err := r.loadAttachments(orderIDs(headers))
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_count":     int64(len(headers)),
		"active_count":    int64(0),
		"expired_count":   int64(0),
		"cancelled_count": int64(0),
		"unbound_count":   int64(0),
	}
	for index := range headers {
		record := buildShipmentRecord(&headers[index], attachments[headers[index].ID])
		switch record.Status {
		case "active":
			stats["active_count"] = stats["active_count"].(int64) + 1
		case "cancelled":
			stats["cancelled_count"] = stats["cancelled_count"].(int64) + 1
		default:
			stats["expired_count"] = stats["expired_count"].(int64) + 1
		}
		if !record.RecordBound {
			stats["unbound_count"] = stats["unbound_count"].(int64) + 1
		}
	}
	return stats, nil
}

// UpsertDetailsForOrder is the only write path for this table. It is called
// explicitly from the warranty/after-sales area after an operator chooses to
// attach evidence to an existing shipped order.
func (r *ShipmentRecordRepository) UpsertDetailsForOrder(
	orderID uint,
	note string,
	images []string,
	productCodes []string,
	warrantyMonths int,
	warrantyStartAt time.Time,
	warrantyExpires time.Time,
) (*shippingdomain.ShipmentRecord, error) {
	orderRecord, err := r.findShippedOrderByID(orderID, true)
	if err != nil {
		return nil, err
	}

	if warrantyMonths <= 0 {
		warrantyMonths = defaultShipmentWarrantyMonths
	}
	if warrantyStartAt.IsZero() {
		warrantyStartAt = orderShippedAt(orderRecord)
	}
	if warrantyExpires.IsZero() {
		warrantyExpires = warrantyStartAt.AddDate(0, warrantyMonths, 0)
	}

	imagesJSON, err := marshalStringList(images)
	if err != nil {
		return nil, err
	}
	productCodesJSON, err := marshalStringList(productCodes)
	if err != nil {
		return nil, err
	}
	itemsJSON, err := json.Marshal(orderRecord.Items)
	if err != nil {
		return nil, err
	}

	customerName := orderCustomerName(orderRecord)
	trackingNumber := strings.TrimSpace(orderRecord.TrackingNumber)
	_, err = r.findAttachmentByOrderID(orderID)
	if err != nil && !IsRecordNotFound(err) {
		return nil, err
	}

	status := "active"
	if !warrantyExpires.After(time.Now()) {
		status = "expired"
	}

	sourceFields := map[string]interface{}{
		"order_number":      orderRecord.OrderNumber,
		"user_id":           orderRecord.UserID,
		"customer_name":     customerName,
		"customer_email":    orderCustomerEmail(orderRecord),
		"tracking_number":   trackingNumber,
		"shipped_at":        orderShippedAt(orderRecord),
		"items_snapshot":    itemsJSON,
		"product_codes":     productCodesJSON,
		"details_bound":     true,
		"shipping_note":     strings.TrimSpace(note),
		"shipping_images":   imagesJSON,
		"warranty_months":   warrantyMonths,
		"warranty_start_at": warrantyStartAt,
		"warranty_expires":  warrantyExpires,
		"status":            status,
	}

	if IsRecordNotFound(err) {
		record := &shippingdomain.ShipmentRecord{
			OrderID:         orderRecord.ID,
			OrderNumber:     orderRecord.OrderNumber,
			UserID:          orderRecord.UserID,
			CustomerName:    customerName,
			CustomerEmail:   orderCustomerEmail(orderRecord),
			TrackingNumber:  trackingNumber,
			ShippedAt:       orderShippedAt(orderRecord),
			ItemsSnapshot:   datatypes.JSON(itemsJSON),
			ProductCodes:    productCodesJSON,
			DetailsBound:    true,
			ShippingNote:    strings.TrimSpace(note),
			ShippingImages:  imagesJSON,
			WarrantyMonths:  warrantyMonths,
			WarrantyStartAt: warrantyStartAt,
			WarrantyExpires: warrantyExpires,
			Status:          status,
		}
		if err := r.db.Create(record).Error; err != nil {
			return nil, err
		}
	} else if err := r.db.Model(&shippingdomain.ShipmentRecord{}).
		Where("order_id = ?", orderID).
		Updates(sourceFields).Error; err != nil {
		return nil, err
	}

	return r.FindByOrderID(orderID)
}

func (r *ShipmentRecordRepository) findAttachmentByOrderID(orderID uint) (*shippingdomain.ShipmentRecord, error) {
	var attachment shippingdomain.ShipmentRecord
	if err := r.db.Where("order_id = ?", orderID).First(&attachment).Error; err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *ShipmentRecordRepository) findShippedOrderByID(orderID uint, withItems bool) (*orderdomain.Order, error) {
	var orderRecord orderdomain.Order
	query := shippedOrderScope(r.db.Model(&orderdomain.Order{})).
		Where("orders.id = ?", orderID)
	if withItems {
		query = query.Preload("Items")
	}
	if err := query.First(&orderRecord).Error; err != nil {
		return nil, err
	}
	return &orderRecord, nil
}

func (r *ShipmentRecordRepository) findShippedOrderByNumber(orderNumber string, userID uint) (*orderdomain.Order, error) {
	var orderRecord orderdomain.Order
	query := shippedOrderScope(r.db.Model(&orderdomain.Order{})).
		Preload("Items").
		Where("orders.order_number = ?", strings.TrimSpace(orderNumber))
	if userID > 0 {
		query = query.Where("orders.user_id = ?", userID)
	}
	if err := query.First(&orderRecord).Error; err != nil {
		return nil, err
	}
	return &orderRecord, nil
}

func (r *ShipmentRecordRepository) findShippedOrderHeaders(keyword string) ([]orderdomain.Order, error) {
	var orders []orderdomain.Order
	query := shippedOrderScope(r.db.Model(&orderdomain.Order{}))
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Where(
			"(LOWER(orders.order_number) LIKE ? OR LOWER(orders.shipping_first_name) LIKE ? OR LOWER(orders.shipping_last_name) LIKE ? OR LOWER(orders.shipping_email) LIKE ? OR LOWER(orders.tracking_number) LIKE ? OR EXISTS (SELECT 1 FROM shipment_records sr WHERE sr.order_id = orders.id AND (LOWER(sr.tracking_number) LIKE ? OR LOWER(CAST(sr.product_codes AS TEXT)) LIKE ?)))",
			like, like, like, like, like, like, like,
		)
	}
	if err := query.
		Order("COALESCE(orders.shipped_at, orders.updated_at, orders.created_at) DESC").
		Order("orders.id DESC").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *ShipmentRecordRepository) loadAttachments(ids []uint) (map[uint]*shippingdomain.ShipmentRecord, error) {
	result := make(map[uint]*shippingdomain.ShipmentRecord, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var attachments []shippingdomain.ShipmentRecord
	if err := r.db.Where("order_id IN ?", ids).Find(&attachments).Error; err != nil {
		return nil, err
	}
	for index := range attachments {
		result[attachments[index].OrderID] = &attachments[index]
	}
	return result, nil
}

func (r *ShipmentRecordRepository) loadOrdersWithItems(ids []uint) (map[uint]*orderdomain.Order, error) {
	result := make(map[uint]*orderdomain.Order, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var orders []orderdomain.Order
	if err := r.db.Preload("Items").Where("id IN ?", ids).Find(&orders).Error; err != nil {
		return nil, err
	}
	for index := range orders {
		result[orders[index].ID] = &orders[index]
	}
	return result, nil
}

func shippedOrderScope(query *gorm.DB) *gorm.DB {
	return query.Where(
		"(orders.shipped_at IS NOT NULL OR orders.status IN ? OR orders.shipping_status IN ?)",
		[]string{"shipped", "completed"},
		[]string{"shipped", "delivered"},
	)
}

func buildShipmentRecord(orderRecord *orderdomain.Order, attachment *shippingdomain.ShipmentRecord) *shippingdomain.ShipmentRecord {
	shippedAt := orderShippedAt(orderRecord)
	itemsJSON, _ := json.Marshal(orderRecord.Items)
	warrantyStart := shippedAt
	warrantyMonths := defaultShipmentWarrantyMonths
	warrantyExpires := warrantyStart.AddDate(0, warrantyMonths, 0)
	status := "active"
	shippingImages := datatypes.JSON([]byte("[]"))
	productCodes := datatypes.JSON([]byte("[]"))
	note := ""
	var trackingShipmentID *uint
	recordBound := false

	if attachment != nil {
		trackingShipmentID = attachment.TrackingShipmentID
		recordBound = attachment.DetailsBound
		note = attachment.ShippingNote
		shippingImages = nonEmptyJSON(attachment.ShippingImages)
		productCodes = nonEmptyJSON(attachment.ProductCodes)
		if attachment.WarrantyMonths > 0 {
			warrantyMonths = attachment.WarrantyMonths
		}
		if !attachment.WarrantyStartAt.IsZero() {
			warrantyStart = attachment.WarrantyStartAt
		}
		if !attachment.WarrantyExpires.IsZero() {
			warrantyExpires = attachment.WarrantyExpires
		} else {
			warrantyExpires = warrantyStart.AddDate(0, warrantyMonths, 0)
		}
		status = attachment.Status
		if status == "" {
			status = "active"
		}
	}

	if status != "cancelled" {
		if warrantyExpires.After(time.Now()) {
			status = "active"
		} else {
			status = "expired"
		}
	}

	trackingNumber := strings.TrimSpace(orderRecord.TrackingNumber)
	if trackingNumber == "" && attachment != nil {
		trackingNumber = strings.TrimSpace(attachment.TrackingNumber)
	}

	return &shippingdomain.ShipmentRecord{
		// The admin/public attachment API is keyed by order ID, including when
		// the optional shipment_records row does not exist yet.
		ID:                 orderRecord.ID,
		OrderID:            orderRecord.ID,
		OrderNumber:        orderRecord.OrderNumber,
		UserID:             orderRecord.UserID,
		CustomerName:       orderCustomerName(orderRecord),
		CustomerEmail:      orderCustomerEmail(orderRecord),
		TrackingShipmentID: trackingShipmentID,
		TrackingNumber:     trackingNumber,
		ShippedAt:          shippedAt,
		ItemsSnapshot:      datatypes.JSON(itemsJSON),
		ProductCodes:       productCodes,
		ShippingNote:       note,
		ShippingImages:     shippingImages,
		WarrantyMonths:     warrantyMonths,
		WarrantyStartAt:    warrantyStart,
		WarrantyExpires:    warrantyExpires,
		Status:             status,
		RecordBound:        recordBound,
		OrderStatus:        orderRecord.Status,
		ShippingState:      orderRecord.ShippingStatus,
	}
}

func shipmentRecordStatusMatches(record *shippingdomain.ShipmentRecord, status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "all":
		return true
	case "unbound":
		return !record.RecordBound
	case "active":
		return record.Status == "active"
	case "expired":
		return record.Status == "expired"
	case "cancelled":
		return record.Status == "cancelled"
	default:
		return true
	}
}

func orderIDs(orders []orderdomain.Order) []uint {
	ids := make([]uint, 0, len(orders))
	for index := range orders {
		ids = append(ids, orders[index].ID)
	}
	return ids
}

func orderShippedAt(orderRecord *orderdomain.Order) time.Time {
	if orderRecord.ShippedAt != nil && !orderRecord.ShippedAt.IsZero() {
		return orderRecord.ShippedAt.UTC()
	}
	if !orderRecord.UpdatedAt.IsZero() {
		return orderRecord.UpdatedAt.UTC()
	}
	if !orderRecord.CreatedAt.IsZero() {
		return orderRecord.CreatedAt.UTC()
	}
	return time.Now().UTC()
}

func orderCustomerName(orderRecord *orderdomain.Order) string {
	name := strings.TrimSpace(strings.TrimSpace(orderRecord.ShippingAddress.FirstName) + " " + strings.TrimSpace(orderRecord.ShippingAddress.LastName))
	if name == "" {
		name = strings.TrimSpace(strings.TrimSpace(orderRecord.BillingAddress.FirstName) + " " + strings.TrimSpace(orderRecord.BillingAddress.LastName))
	}
	return name
}

func orderCustomerEmail(orderRecord *orderdomain.Order) string {
	email := strings.TrimSpace(orderRecord.ShippingAddress.Email)
	if email == "" {
		email = strings.TrimSpace(orderRecord.BillingAddress.Email)
	}
	return email
}

func marshalStringList(values []string) (datatypes.JSON, error) {
	if values == nil {
		values = []string{}
	}
	raw, err := json.Marshal(values)
	return datatypes.JSON(raw), err
}

func nonEmptyJSON(raw datatypes.JSON) datatypes.JSON {
	if len(raw) == 0 || string(raw) == "null" {
		return datatypes.JSON([]byte("[]"))
	}
	return raw
}
