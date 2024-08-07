import os
import requests
import time
import random
from fake_useragent import UserAgent
from concurrent.futures import ThreadPoolExecutor, as_completed

directory = "./file/original"
if not os.path.exists(directory):
    os.makedirs(directory)

start_number = 1000
end_number = 3000 # max: 154999

url_template = "https://www.cac.edu.tw/apply113/system/ColQry_vforStu113apply_GF84ad9zx/html/113_{}.htm?v=1.0"  # 請根據實際情況調整URL

ua = UserAgent()

max_retry_attempts = 3

def download_file(number):
    file_number = f"{number:06}"
    url = url_template.format(file_number)

    headers = {
        'User-Agent': ua.random,
        'Accept-Language': 'en-US,en;q=0.9',
        'Accept-Encoding': 'gzip, deflate',
        'Connection': 'keep-alive',
    }
    
    retry = 0

    delay = random.uniform(1, 5)
    time.sleep(delay)

    while retry < max_retry_attempts:
        try:
            response = requests.get(url, headers=headers)
            
            if response.status_code == 200:
                file_path = os.path.join(directory, f"{file_number}.html")
                
                with open(file_path, "wb") as file:
                    file.write(response.content)
                    
                print(f"文件 {file_number}.html 下載成功")
                return
            else:
                print(f"文件 {file_number}.html 不存在 (狀態碼: {response.status_code})")
                return
        
        except requests.RequestException as e:
            print(f"下載文件 {file_number}.html 時發生錯誤: {e}")
        
        retry += 1
        delay = random.uniform(1, 5)
        time.sleep(delay)
    
    print(f"文件 {file_number}.html 下載失敗，已達到最大重試次數")

def main():
    with ThreadPoolExecutor(max_workers=50) as executor:
        futures = [executor.submit(download_file, number) for number in range(start_number, end_number + 1)]
        
        for future in as_completed(futures):
            future.result()

if __name__ == "__main__":
    main()
