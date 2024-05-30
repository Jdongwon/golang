package main

import (
    "bufio"
    "encoding/csv"
    "fmt"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "time"
)

// 시스템 정보를 저장하는 구조체
type TopInfo struct {
    Time    string
    CPUUtil string
    MemUtil string
}

// 파일 존재 여부를 확인하는 함수
func top_checkFileExists(filePath string) bool {
    _, err := os.Stat(filePath)	//os.Stat - 파일, 디렉터리 정보를 가져옴, 파일이 존재하지 않거나 접근 불가 시 에러 반환
    return err == nil	//파일이 존재하면 True, 존재하지 않으면 False 반환
}

// top 명령어 출력을 파싱하여 시스템 정보를 추출하는 함수
func top_parseOutput(output string) ([]TopInfo, error) {
    scanner := bufio.NewScanner(strings.NewReader(output))
		// 문자열을 한 줄씩 읽을 수 있도록 설정하는 부분
		// strings.NewReader - 문자열 입력받아 'io.Reader' 인터페이스를 구현하는 객체를 생성
		// bufio.NewScanner - 'io.Reader'를 입력받아 데이터를 한 줄씩 읽을 수 있는 스캐너 생성
    var top_infos []TopInfo
    var cpuUtil, memUtil, total, free float64

    for scanner.Scan() {
		// scanner.Scan() - 한 줄씩 읽음
        line := scanner.Text()
		// scanner.Text() - 해당 줄의 내용을 문자열로 반환
		// line에 스캐너가 읽은 한 줄의 텍스트
        if strings.Contains(line, "%Cpu(s):") {
            fields := strings.Fields(line)
				// strings.Fields() - 문자열을 공백을 기준으로 분리하여 필드의 슬라이스([]string)로 반환
				// 예를 들어, %Cpu(s): 1.2 us, 2.3 sy, 0.0 ni, 95.4 id, 0.1 wa..."라는 줄을 공백 기준으로 분리
            for i, field := range fields {
					// i = index, field = 값
                if field == "id," {
                    idle, err := strconv.ParseFloat(fields[i-1], 64)
						// fields[i-1] - id 값 바로 앞의 값을 반환
						// strconv.ParseFloat( , 64) - float 64로 변환
                    if err != nil {
                        return top_infos, err
                    }
                    cpuUtil = 100 - idle
                    break
                }
            }
        } else if strings.Contains(line, "MiB Mem") {
            fields := strings.Fields(line)

            for i, field := range fields {
                if field == "total," {
					var err error
                    total, err = strconv.ParseFloat(fields[i-1], 64)
                    if err != nil {
                        return top_infos, err
                    }
                } else if field == "free," {
					var err error
                    free, err = strconv.ParseFloat(fields[i-1], 64)
                    if err != nil {
                        return top_infos, err
                    }
                }
            }
            memUtil = (total - free) / total * 100
        }
    }

    now := time.Now().Format("2006-01-02 15:04:05")
    top_info := TopInfo{
        Time:    now,
        CPUUtil: fmt.Sprintf("%.2f", cpuUtil),
        MemUtil: fmt.Sprintf("%.2f", memUtil),
    }
    top_infos = append(top_infos, top_info)
    return top_infos, scanner.Err()
}

// 시스템 정보를 CSV 파일에 저장하는 함수
func top_writeToCSV(writer *csv.Writer, top_infos []TopInfo) error {
    for _, info := range top_infos {
        record := []string{
            info.Time,
            info.CPUUtil,
            info.MemUtil,
        }
        writer.Write(record)
    }
    writer.Flush()
    return nil
}

func main() {
    // CSV 파일 경로
    filePath := "/home/noah/project/top_check/result.csv"
    fileExists := top_checkFileExists(filePath)

    // CSV 파일 열기
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("CSV 파일 열기 오류:", err)
        return
    }
    defer file.Close()

    writer := csv.NewWriter(file)

    // 파일이 처음 생성되었을 때 헤더 추가
    if !fileExists {
        header := []string{"Time", "CPU Util", "Memory Util"}

        if err := writer.Write(header); err != nil {
            fmt.Println("CSV 헤더 쓰기 오류:", err)
        }

        writer.Flush() // 헤더를 기록 후 버퍼를 플러시합니다.
    }

    for {
        // top 명령어 실행
        top_cmd := exec.Command("top", "-bn1")
        top_out, err := top_cmd.Output()
        if err != nil {
            fmt.Println("top 실행 오류:", err)
            return
        }

        // top 명령어 출력 파싱
        top_infos, err := top_parseOutput(string(top_out))
        if err != nil {
            fmt.Println("top 출력 파싱 오류:", err)
            return
        }

        // CSV 파일에 데이터 쓰기
        err = top_writeToCSV(writer, top_infos)
        if err != nil {
            fmt.Println("CSV 파일 쓰기 오류:", err)
            return
        }

        // 1초 대기
        time.Sleep(1 * time.Second)
    }
}
