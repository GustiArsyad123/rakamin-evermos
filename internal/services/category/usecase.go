package category

import "github.com/example/ms-ecommerce/internal/pkg/models"

type Usecase interface {
	Create(name string) (int64, error)
	Update(id int64, name string) error
	Delete(id int64) error
	Get(id int64) (*models.Category, error)
	List(filters map[string]string, page, limit int) ([]*models.Category, int, error)
}

type categoryUsecase struct {
	repo Repository
}

func NewUsecase(r Repository) Usecase {
	return &categoryUsecase{repo: r}
}

func (u *categoryUsecase) Create(name string) (int64, error) {
	return u.repo.Create(name)
}

func (u *categoryUsecase) Update(id int64, name string) error {
	return u.repo.Update(id, name)
}

func (u *categoryUsecase) Delete(id int64) error {
	return u.repo.Delete(id)
}

func (u *categoryUsecase) Get(id int64) (*models.Category, error) {
	return u.repo.GetByID(id)
}

func (u *categoryUsecase) List(filters map[string]string, page, limit int) ([]*models.Category, int, error) {
	return u.repo.List(filters, page, limit)
}
