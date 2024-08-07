import argparse
import pymysql
import os
import csv

# 設定命令列參數解析
parser = argparse.ArgumentParser(description='Insert CSV data into MySQL database')
parser.add_argument('-db_user', type=str, required=True, help='Database user')
parser.add_argument('-db_password', type=str, required=True, help='Database password')
parser.add_argument('-db_host', type=str, required=True, help='Database host')
parser.add_argument('-db_port', type=int, required=True, help='Database port')
parser.add_argument('-db_name', type=str, required=True, help='Database name')

args = parser.parse_args()

# 設定資料庫連接
db = pymysql.connect(
    host=args.db_host,
    user=args.db_user,
    password=args.db_password,
    database=args.db_name,
    port=args.db_port
)

cursor = db.cursor()

# CSV目錄設置
csv_directory = "./file/csv/112_"

# 遍歷CSV檔案並處理每個檔案
for filename in os.listdir(csv_directory):
    if filename.endswith('.csv'):
        file_path = os.path.join(csv_directory, filename)
        
        # 讀取CSV檔案
        with open(file_path, 'r', encoding='utf-8') as csv_file:
            reader = csv.reader(csv_file)
            rows = list(reader)
            
            if rows:
                # 取CSV檔案中的最後一筆資料
                last_row = rows[-1]
                department_code = last_row[0]
                minimum_standard = last_row[5]
                
                # 準備資料庫更新或插入的SQL語句
                update_sql = """
                UPDATE 112_year
                SET candidates = %s
                WHERE code = %s
                """
                
                # 更新或插入資料
                cursor.execute("SELECT COUNT(*) FROM 112_year WHERE code = %s", (department_code,))
                exists = cursor.fetchone()[0]
                
                if exists:
                    cursor.execute(update_sql, (minimum_standard, department_code))
                else:
                    print(department_code)
                
                # 提交更改
                db.commit()

# 關閉資料庫連接
cursor.close()
db.close()
