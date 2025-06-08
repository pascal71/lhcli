// pkg/client/stubs.go
package client

import (
	"context"
	"fmt"
)

// volumeClient implements VolumeInterface
type volumeClient struct {
	client *Client
}

func (v *volumeClient) List() ([]Volume, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (v *volumeClient) Get(name string) (*Volume, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (v *volumeClient) Create(volume *VolumeCreateInput) (*Volume, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (v *volumeClient) Delete(name string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (v *volumeClient) Update(name string, volume *VolumeUpdateInput) (*Volume, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (v *volumeClient) Attach(name string, input *VolumeAttachInput) (*Volume, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (v *volumeClient) Detach(name string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

// settingsClient implements SettingsInterface
type settingsClient struct {
	client *Client
}

// backupClient implements BackupInterface
type backupClient struct {
	client *Client
}

func (b *backupClient) List(volumeName string) ([]Backup, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (b *backupClient) Get(backupName string) (*Backup, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (b *backupClient) Create(volumeName string, input *BackupCreateInput) (*Backup, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (b *backupClient) Delete(backupName string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (b *backupClient) GetTarget() (*BackupTarget, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (b *backupClient) SetTarget(target *BackupTarget) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

// engineImageClient implements EngineImageInterface
type engineImageClient struct {
	client *Client
}

// eventClient implements EventInterface
type eventClient struct {
	client *Client
}

func (e *eventClient) List(opts EventListOptions) ([]Event, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (e *eventClient) Watch(
	ctx context.Context,
	opts EventListOptions,
	callback func(Event),
) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}
