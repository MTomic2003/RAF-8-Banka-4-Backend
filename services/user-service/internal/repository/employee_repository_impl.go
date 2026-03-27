package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, employee *model.Employee) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("ActuaryInfo").Create(employee).Error; err != nil {
			return err
		}

		return syncActuaryInfo(tx, employee)
	})
}

func (r *employeeRepository) FindByID(ctx context.Context, id uint) (*model.Employee, error) {
	var e model.Employee
	result := r.db.WithContext(ctx).
		Preload("Permissions").
		Preload("ActuaryInfo").
		Preload("Identity").
		First(&e, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &e, nil
}

func (r *employeeRepository) FindByIdentityID(ctx context.Context, identityID uint) (*model.Employee, error) {
	var e model.Employee
	result := r.db.WithContext(ctx).
		Preload("Permissions").
		Preload("ActuaryInfo").
		Preload("Identity").
		Where("identity_id = ?", identityID).
		First(&e)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &e, nil
}

func (r *employeeRepository) Update(ctx context.Context, employee *model.Employee) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Permissions", "ActuaryInfo").Save(employee).Error; err != nil {
			return err
		}

		if err := tx.Where("employee_id = ?", employee.EmployeeID).Delete(&model.EmployeePermission{}).Error; err != nil {
			return err
		}

		if len(employee.Permissions) > 0 {
			if err := tx.Create(&employee.Permissions).Error; err != nil {
				return err
			}
		}

		return syncActuaryInfo(tx, employee)
	})
}

func (r *employeeRepository) GetAll(ctx context.Context, email, firstName, lastName, position string, page, pageSize int) ([]model.Employee, int64, error) {
	var employees []model.Employee
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.Employee{}).
		Preload("Position").
		Preload("Permissions").
		Preload("ActuaryInfo").
		Preload("Identity").
		Joins("LEFT JOIN positions ON positions.position_id = employees.position_id").
		Joins("LEFT JOIN identities ON identities.id = employees.identity_id")

	if email != "" {
		query = query.Where("identities.email ILIKE ?", "%"+email+"%")
	}
	if firstName != "" {
		query = query.Where("employees.first_name ILIKE ?", "%"+firstName+"%")
	}
	if lastName != "" {
		query = query.Where("employees.last_name ILIKE ?", "%"+lastName+"%")
	}
	if position != "" {
		query = query.Where("positions.title ILIKE ?", "%"+position+"%")
	}

	if err := query.Model(&model.Employee{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("employees.employee_id DESC").Find(&employees).Error; err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

func syncActuaryInfo(tx *gorm.DB, employee *model.Employee) error {
	if employee.ActuaryInfo == nil || !employee.ActuaryInfo.HasRole() {
		return tx.Where("employee_id = ?", employee.EmployeeID).Delete(&model.ActuaryInfo{}).Error
	}

	actuary := *employee.ActuaryInfo
	actuary.EmployeeID = employee.EmployeeID

	return tx.Save(&actuary).Error
}
