import os
import json
import pymysql
import argparse

# 設定命令列參數解析
parser = argparse.ArgumentParser(description='Insert JSON files data into MySQL database')
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

# 設定JSON檔案的目錄
json_dir = 'file/json'

# 遍歷目錄中的所有檔案
for filename in os.listdir(json_dir):
    if filename.endswith('.json'):
        file_path = os.path.join(json_dir, filename)
        
        with open(file_path, 'r', encoding='utf-8') as f:
            data = json.load(f)
            
            # 提取需要的欄位
            tag = data.get('tag')
            code = data.get('code')
            school_name = data.get('school_name')
            department_name = data.get('department_name')
            combined_name = f"{school_name} {department_name}"
            enrollment_quota = data.get('enrollment_quota')
            screening_date = data.get('screening_date')
            subject = json.dumps(data.get('subject', []))  # 將列表轉換為JSON格式字串
            subject_criteria = json.dumps(data.get('subject_criteria', []))  # 將列表轉換為JSON格式字串
            full_json = json.dumps(data)  # 將整個JSON轉換為字串
            
            # 插入資料到資料庫
            sql = """
            INSERT INTO 113_year (tag, code, combined_name, enrollment_quota, screening_date, subject, subject_criteria, full_json)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
            """
            cursor.execute(sql, (tag, code, combined_name, enrollment_quota, screening_date, subject, subject_criteria, full_json))
            db.commit()

# 關閉資料庫連接
cursor.close()
db.close()
