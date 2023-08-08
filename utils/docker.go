package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// Global variables
var (
	// Docker client
	cli *client.Client
	// Docker context
	ctx = context.Background()
)

// Connects to the Docker daemon over TLS via client.crt and client.key.
func ConnectToDocker() error {
	// Loading environment variables
	CERT_PATH := os.Getenv("CERT_PATH")
	DOCKER_HOST := os.Getenv("DOCKER_HOST")
	DOCKER_PORT := os.Getenv("DOCKER_PORT")

	// Load the client certificate and key.
	cert, err := tls.LoadX509KeyPair(filepath.Join(CERT_PATH, "client.crt"), filepath.Join(CERT_PATH, "client.key"))
	if err != nil {
		return fmt.Errorf("failed to load client certificate and key: %v", err)
	}

	// Load the CA certificate.
	caCert, err := ioutil.ReadFile(filepath.Join(CERT_PATH, "ca.crt"))
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create a new TLS configuration with the client certificate, key, and CA certificate.
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	// Add the client certificate to the TLS configuration.
	tlsConfig.BuildNameToCertificate()

	// Create a new transport with the TLS configuration.
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Create a new HTTP client with the transport.
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 10,
	}

	// Create a new Docker client with the HTTP client.
	cli, err = client.NewClientWithOpts(
		client.WithHTTPClient(httpClient),
		client.WithAPIVersionNegotiation(),
		client.WithHost("tcp://"+DOCKER_HOST+":"+DOCKER_PORT),
		client.WithTLSClientConfig(filepath.Join(CERT_PATH, "ca.crt"), filepath.Join(CERT_PATH, "client.crt"), filepath.Join(CERT_PATH, "client.key")),
	)
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}

	// Ping the Docker daemon to test the connection.
	_, err = cli.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping Docker daemon: %v", err)
	}

	return err
}

// Lists the containers on the Docker daemon.
func ListContainers() ([]types.Container, error) {
	// List the containers.
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	return containers, nil
}

// Gets the stats of a container by ID.
func GetContainerStatsByID(containerID string) (*types.StatsJSON, error) {
	// Get the stats of the container with the specified ID.
	stats, err := cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return &types.StatsJSON{}, fmt.Errorf("failed to get stats for container %s: %v", containerID, err)
	}
	defer stats.Body.Close()

	// Decode the JSON response into a StatsJSON object.
	var containerStats types.StatsJSON
	decoder := json.NewDecoder(stats.Body)
	if err := decoder.Decode(&containerStats); err != nil {
		return &types.StatsJSON{}, fmt.Errorf("failed to decode stats for container %s: %v", containerID, err)
	}

	return &containerStats, nil
}

// Starts a container by ID.
func StartContainerByID(containerID string) error {
	err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container %s: %v", containerID, err)
	}

	return nil
}

// Stops a container by ID.
func StopContainerByID(containerID string) error {
	timeout := int(time.Second * 10)

	// Create a new ContainerStopOptions value with the specified timeout.
	stopOpts := container.StopOptions{
		Timeout: &timeout,
	}

	err := cli.ContainerStop(ctx, containerID, stopOpts)
	if err != nil {
		return fmt.Errorf("failed to stop container %s: %v", containerID, err)
	}

	return nil
}

// Updates a container with the latest image
func UpdateContainerByID(containerID string) error {
	// Get the existing container configuration.
	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container %s info: %v", containerID, err)
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: containerInfo.NetworkSettings.Networks,
	}

	// Pull the latest image.
	out, err := cli.ImagePull(ctx, containerInfo.Config.Image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %v", containerInfo.Config.Image, err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	// Create a new container using the existing configuration.
	resp, err := cli.ContainerCreate(ctx, containerInfo.Config, containerInfo.HostConfig, networkConfig, nil, containerInfo.Name)
	if err != nil {
		return fmt.Errorf("failed to create container from image %s: %v", containerInfo.Config.Image, err)
	}

	// Start the new container.
	if err := StartContainerByID(resp.ID); err != nil {
		return err
	}

	// Remove the old container.
	if err := DeleteContainerByID(containerID); err != nil {
		return err
	}

	return nil
}

// Deletes a container by ID.
func DeleteContainerByID(containerID string) error {
	// Remove the container.
	if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("failed to remove container %s: %v", containerID, err)
	}

	return nil
}
