package server_group

import (
	"errors"

	"gosee/internal/model"
	"gosee/internal/repository"
)

// Service 服务器分组服务
type Service struct {
	repo *repository.ServerGroupRepository
}

func NewService(repo *repository.ServerGroupRepository) *Service {
	return &Service{repo: repo}
}

// SaveInput 创建/编辑分组入参
type SaveInput struct {
	ID     int64  `json:"id"`
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
}

func (s *Service) List(keyword string) ([]model.ServerGroup, error) {
	return s.repo.List(keyword)
}

func (s *Service) Create(in SaveInput) (*model.ServerGroup, error) {
	g := &model.ServerGroup{Name: in.Name, Remark: in.Remark}
	if err := s.repo.Create(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) Update(in SaveInput) (*model.ServerGroup, error) {
	g, err := s.repo.FindByID(in.ID)
	if err != nil {
		return nil, errors.New("分组不存在")
	}
	g.Name = in.Name
	g.Remark = in.Remark
	if err := s.repo.Update(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) Delete(id int64) error {
	count, err := s.repo.CountServers(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该分组下仍有服务器，无法删除")
	}
	return s.repo.Delete(id)
}
