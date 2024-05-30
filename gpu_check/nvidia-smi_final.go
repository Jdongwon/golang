package main

import (
    "bufio"
    "encoding/csv"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"
)

// GPU 정보를 저장하는 구조체 선언
type GPUInfo struct {
    Time         string
    GPUNumber    string
    Temperature  string
    Power        string
    MemoryUsage  string
    GPUUtil      string
}


// 파일 존재 여부 확인
func gpu_checkFileExists(filePath string) bool {
    _, err := os.Stat(filePath)	//os.Stat - 파일, 디렉터리 정보를 가져옴, 파일이 존재하지 않거나 접근 불가 시 에러 반환
    return err == nil	//파일이 존재하면 True, 존재하지 않으면 False 반환
}

// nvidia-smi 명령어 출력을 파싱하여 GPU 정보를 추출
func gpu_parseOutput(output string) ([]GPUInfo, error) {
    var gpuInfos []GPUInfo
    scanner := bufio.NewScanner(strings.NewReader(output))
		// 문자열을 한 줄씩 읽을 수 있도록 설정하는 부분
		// strings.NewReader - 문자열 입력받아 'io.Reader' 인터페이스를 구현하는 객체를 생성
		// bufio.NewScanner - 'io.Reader'를 입력받아 데이터를 한 줄씩 읽을 수 있는 스캐너 생성

    for scanner.Scan() {
        // scanner.Scan() - 한 줄씩 읽음
		line := scanner.Text()
		// scanner.Text() - 해당 줄의 내용을 문자열로 반환
		// line에 스캐너가 읽은 한 줄의 텍스트
        parts := strings.Split(line, ", ")
		// ', ' 콤마와 공백을 기준으로 문자열을 분리하여 문자열 슬라이스로 반환
			// line := "0, 55, 100, 2000, 30"
			// parts := strings.Split(line, ", ")
			// parts := []string{"0", "55", "100", "2000", "30"}
        if len(parts) != 5 {
            continue
        }
			// 슬라이스 길이는 5가 아니면 다음 반복으로 넘어감
			// --query-gpu에서 받아온 5개의 정보(gpu index, power, memory 등)

        now := time.Now().Format("2006-01-02 15:04:05")
			// UTC 표준 시간대로 현재 시간을 가져오고, RFC3339 형식의 문자열로 변환
        gpuInfo := GPUInfo{
            Time:        now,
            GPUNumber:   parts[0],
            Temperature: parts[1],
            Power:       parts[2],
            MemoryUsage: parts[3],
            GPUUtil:     parts[4],
        }
        gpuInfos = append(gpuInfos, gpuInfo)
		// gpuInfos에 gpuInfo를 append
    }

    if err := scanner.Err(); err != nil {
        return gpuInfos, err
    }

    return gpuInfos, nil
}

// GPU 정보를 CSV 파일에 저장
func gpu_writeToCSV(writer *csv.Writer, gpuInfos []GPUInfo) error {
		// csv.Writer는 패키지에서 정의된 구조체
		// defer writer.Flush()를 정상적으로 동작하기 위해 포인터로 구조체 가져옴
    for _, info := range gpuInfos {
        record := []string{
            info.Time,
            info.GPUNumber,
            info.Temperature,
            info.Power,
            info.MemoryUsage,
            info.GPUUtil,
        }
        if err := writer.Write(record); err != nil {
            return err
        }
			// csv에 기록함
    }
    writer.Flush()
    return writer.Error()
}



func main() {
    // CSV 파일이 없으면 헤더를 추가하여 생성
    filePath := "/home/noah/project/gpu_check/result.csv"
    fileExists := gpu_checkFileExists(filePath)

    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		// os.OpenFile - 파일을 열거나, 파일이 존재하지 않으면 새로 생성
		// os.O_APPEND - 파일이 존재할 경우, 파일 끝에 데이터 추가
		// os.O_CREATE - 파일이 존재하지 않으면 새로 생성
		// os.O_WRONLY - 파일을 쓰기 전용으로 오픈
		// 0644 - 소유자에게 읽기, 쓰기 권한 / 그룹, 다른 사용자에게 읽기 권한 부여
		// file은 os.OpenFile함수로 열거나 생성하는 파일 핸들 
    if err != nil {
        fmt.Println("CSV 파일 열기 오류:", err)
        return
    }
    defer file.Close()
		// main 함수가 끝나면 파일을 닫음
		// defer - main 함수가 끝나면 해당 기능이 실행되도록 함, defer가 여러개 일 경우 LIFO, 즉 file.Close는 가장 마지막에 실행됨
    writer := csv.NewWriter(file)
		// csv 파일에 데이터 입력을 위한 메서드 제공
		// writer.Write(header) - 헤더 행을 csv 파일에 입력
		// writer.Write(record) - 하나의 레코드(데이터)를 csv 파일에 입력
    defer writer.Flush()
		// 버퍼(메모리)에 저장된 데이터를 실제 파일에 기록

    // 파일이 처음 생성되었을 때 헤더 추가
    if !fileExists {
		//fileExists가 True면 동작 x, fileExists가 False면 헤더 추가
        header := []string{"Time", "GPU Number", "GPU Temperature(℃)", "Power(W)", "Memory-Usage(MiB)", "GPU-Util(%)"}
        
        if err := writer.Write(header); err != nil {
            fmt.Println("CSV 헤더 쓰기 오류:", err)
        }

        writer.Flush()
    }

    for {
        gpu_cmd := exec.Command("nvidia-smi", "--query-gpu=index,temperature.gpu,power.draw,memory.used,utilization.gpu", "--format=csv,noheader,nounits")
			// exec.Command("nvidia-smi") - nvidia-smi를 실행할 수 있도록 준비
			// --query-gpu - nvidia-smi에서 정보를 쿼리할 수 있도록 미리 설정된 옵션
			// --format=csv,noheader,nounits - 출력 형식을 csv 설정, 헤더 포함x, 단위 포함x
        gpu_out, err := gpu_cmd.Output()
			// cmd.Output() - 명령 실행 및 출력 반환
			// cmd.Run() - 명령 실행 및 완료될 때까지 기다림
			// cmd.CombinedOutput() - 명령 실행 및 출력과 오류를 결합하여 반환
			// cmd.Start() - 명령어를 비동기적으로 시작, 즉 명령어가 시작된 후 곧바로 반환
			// cmd.Wait() - 명령어가 완료될 때까지 기다림
        if err != nil {
            fmt.Println("nvidia-smi 실행 오류:", err)
            return
        }

        gpuInfos, err := gpu_parseOutput(string(gpu_out))
        if err != nil {
            fmt.Println("nvidia-smi 출력 파싱 오류:", err)
            return
        }

        if err := gpu_writeToCSV(writer, gpuInfos); err != nil {
            fmt.Println("CSV 파일 쓰기 오류:", err)
            return
        }

        time.Sleep(1 * time.Second)
			// 1초 대기 후 다시 for문 실행
    }
}

