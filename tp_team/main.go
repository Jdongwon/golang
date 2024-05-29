package main

import (
    "bufio"
    "encoding/csv"
    "fmt"
    "os"
    "os/exec"
    "os/signal"
    "path/filepath"
    "strings"
    "syscall"
    "time"
)

// GPU 정보를 저장하는 구조체
type GPUInfo struct {
    Time         string
    GPUNumber    string
    Temperature  string
    Power        string
    MemoryUsage  string
    GPUUtil      string
}

func main() {
    // 홈 디렉토리에 파일 저장
    homeDir, err := os.UserHomeDir()
    if err != nil {
        fmt.Println("홈 디렉토리를 찾는 중 오류 발생:", err)
        return
    }
    filePath := filepath.Join(homeDir, "gpu_stress_test_result.csv")

    fileExists := checkFileExists(filePath)

    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("CSV 파일 열기 오류:", err)
        return
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // 파일이 처음 생성되었을 때 헤더 추가
    if !fileExists {
        header := []string{"Time", "GPU Number", "GPU Temperature(℃)", "Power(W)", "Memory-Usage(MiB)", "GPU-Util(%)"}
        writer.Write(header)
    }

    // 인터럽트 시그널 처리
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Printf("\n데이터가 파일에 저장되었습니다: %s\n", filePath)
        os.Exit(0)
    }()

    for {
        cmd := exec.Command("nvidia-smi", "--query-gpu=index,temperature.gpu,power.draw,memory.used,utilization.gpu", "--format=csv,noheader,nounits")
        stdout, err := cmd.Output()
        if err != nil {
            fmt.Println("nvidia-smi 실행 오류:", err)
            return
        }

        gpuInfos := parseOutput(string(stdout))
        writeToCSV(writer, gpuInfos)

        time.Sleep(60 * time.Second) // 1분(60초)마다 실행
    }
}

// 파일 존재 여부 확인
func checkFileExists(filePath string) bool {
    _, err := os.Stat(filePath)
    return err == nil
}

// nvidia-smi 명령어 출력을 파싱하여 GPU 정보를 추출
func parseOutput(output string) []GPUInfo {
    var gpuInfos []GPUInfo
    scanner := bufio.NewScanner(strings.NewReader(output))

    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, ", ")
        if len(parts) != 5 {
            continue
        }

        now := time.Now().UTC().Format(time.RFC3339)
        gpuInfo := GPUInfo{
            Time:        now,
            GPUNumber:   parts[0],
            Temperature: parts[1],
            Power:       parts[2],
            MemoryUsage: parts[3],
            GPUUtil:     parts[4],
        }
        gpuInfos = append(gpuInfos, gpuInfo)
    }

    return gpuInfos
}

// GPU 정보를 CSV 파일에 저장
func writeToCSV(writer *csv.Writer, gpuInfos []GPUInfo) {
    for _, info := range gpuInfos {
        record := []string{
            info.Time,
            info.GPUNumber,
            info.Temperature,
            info.Power,
            info.MemoryUsage,
            info.GPUUtil,
        }
        writer.Write(record)
    }
    writer.Flush()
}
