package service

import (
	"context"
	"kzhikcn/internal/appinfo"
	"kzhikcn/pkg/data"
	"kzhikcn/server/common/secutils"

	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

var (
	ErrAdminNotFound            = errors.New("admin not found")
	ErrAdminValidateFailed      = errors.New("admin validate failed")
	ErrAdminComparePasswordFail = errors.New("admin compare password failed")
)

type AdminService struct {
	ctx *ServiceContext
}

type AdminUpdateFields struct {
	Username *string
	Email    *string
	Avatar   *string
}

type TOTPResult struct {
	Secret      string `json:"secret"`
	AccountName string `json:"accountName"`
	Issuer      string `json:"issuer"`
	URL         string `json:"url"`
}

func NewAdminService(ctx *ServiceContext) *AdminService {
	return &AdminService{ctx: ctx}
}

func (a *AdminService) GetAdminById(ctx context.Context, adminID uint) (*data.Admin, error) {
	return a.getAdminById(ctx, adminID)
}

func (a *AdminService) GetAdminByName(ctx context.Context, username string) (*data.Admin, error) {
	admins, err := data.GetAdmin(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("username=?", username).Limit(1)
	})
	if admins == nil || errors.Is(err, gorm.ErrRecordNotFound) || len(admins) == 0 {
		return nil, ErrAdminNotFound
	}
	return &admins[0], nil
}

func (a *AdminService) GetProfile(ctx context.Context, adminID uint) (*data.Admin, error) {
	return a.getAdminById(ctx, adminID, func(tx *gorm.DB) *gorm.DB {
		return tx.Omit("password", "totp_secret")
	})
}

func (a *AdminService) VerifyPassword(ctx context.Context, adminID uint, password string) error {
	admin, err := a.getAdminById(ctx, adminID, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("password", "id")
	})
	if err != nil {
		return err
	}

	ok, err := secutils.ComparePassword(admin.Password, password)
	if err != nil {
		return errors.Wrap(ErrAdminComparePasswordFail, err.Error())
	}
	if !ok {
		return ErrAdminValidateFailed
	}
	return nil
}

func (a *AdminService) ChangePassword(ctx context.Context, adminID uint, oldPw, newPw string) error {
	if err := a.VerifyPassword(ctx, adminID, oldPw); err != nil {
		return err
	}

	return data.UpdateAdminByID(adminID, &data.Admin{Password: []byte(newPw)}, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("password")
	})
}

func (a *AdminService) UpdateInfo(ctx context.Context, adminID uint, fields AdminUpdateFields) error {
	selected := []string{}
	admin := &data.Admin{}

	if fields.Avatar != nil {
		admin.Avatar = *fields.Avatar
		selected = append(selected, "avatar")
	}
	if fields.Username != nil {
		admin.Username = *fields.Username
		selected = append(selected, "username")
	}
	if fields.Email != nil {
		admin.Email = *fields.Email
		selected = append(selected, "email")
	}

	if len(selected) == 0 {
		return nil
	}

	return data.UpdateAdminByID(adminID, admin, func(tx *gorm.DB) *gorm.DB {
		return tx.Select(selected)
	})
}

func (a *AdminService) GenerateTOTP(ctx context.Context, adminID uint, password, accountName string) (*TOTPResult, error) {
	if err := a.VerifyPassword(ctx, adminID, password); err != nil {
		return nil, err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      appinfo.Name,
		AccountName: accountName,
	})
	if err != nil {
		return nil, err
	}

	err = data.UpdateAdminByID(adminID, &data.Admin{TotpSecret: []byte(key.Secret())}, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("totp_secret")
	})
	if err != nil {
		return nil, err
	}

	return &TOTPResult{
		Secret:      key.Secret(),
		AccountName: key.AccountName(),
		Issuer:      key.Issuer(),
		URL:         key.URL(),
	}, nil
}

func (a *AdminService) SetMFA(ctx context.Context, adminID uint, enable bool, password, otp string) error {
	admin, err := a.getAdminById(ctx, adminID, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("totp_secret", "password", "id")
	})
	if err != nil {
		return err
	}

	ok, _ := secutils.ComparePassword(admin.Password, password)
	if !ok {
		return ErrAdminValidateFailed
	}

	if !totp.Validate(otp, string(admin.TotpSecret)) {
		return errors.New("invalid otp")
	}

	return data.UpdateAdminByID(adminID, &data.Admin{EnableMFA: enable}, func(tx *gorm.DB) *gorm.DB {
		return tx.Select("enable_mfa")
	})
}

func (a *AdminService) getAdminById(ctx context.Context, adminID uint, mods ...func(tx *gorm.DB) *gorm.DB) (*data.Admin, error) {
	admins, err := data.GetAdmin(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Where("id=?", adminID).Limit(1)
		for _, m := range mods {
			tx = m(tx)
		}
		return tx
	})
	if admins == nil || errors.Is(err, gorm.ErrRecordNotFound) || len(admins) == 0 {
		return nil, ErrAdminNotFound
	}
	return &admins[0], nil
}
