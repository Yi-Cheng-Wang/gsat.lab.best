import re
import json
import os
from datetime import datetime

folder_name = "./file/original/112"
output_folder = "./file/json/112"
log_folder = "./file/log"

def create_json_file(data):
    json_data = json.dumps(data, ensure_ascii=False, indent=4)

    json_file_path = os.path.join(output_folder, f"{data['code']}.json")
    with open(json_file_path, 'w', encoding='utf-8') as json_file:
        json_file.write(json_data)

def format_dates(input_string):
    date_pattern = re.compile(r'(\d{3}\.\d{1,2}\.\d{1,2})')
    
    dates = date_pattern.findall(input_string)
    
    if len(dates) == 1:
        return dates[0]
    elif len(dates) == 2:
        if 'font-size' in input_string:
            return '至'.join(dates)
        elif '<br>' in input_string:
            return '、'.join(dates)
    else:
        return input_string

def general_group(content):
    data = {"tag": "general"}

    pattern = r'<td class="Rb" style="color:#FF0000;width:60px;">(.*?)</td>'
    data["code"] = re.findall(pattern, content)[0]

    pattern = r'<div class="colname">(.*?)</div>'
    data["school_name"] = re.findall(pattern, content)[0]
    
    pattern = r'<div class="gsdname">(.*?)</div>'
    data["department_name"] = re.findall(pattern, content)[0]
    
    pattern = r'<td class="Rb">(.*?)</td>'
    data["enrollment_quota"] = re.findall(pattern, content)[0]
    
    pattern = r'<td class="Rb" style="line-height:12px;">(.*?)</td>'
    data["screening_date"] = format_dates(re.findall(pattern, content)[0])
    
    pattern = r'<td class="Bb font_bold" colspan="2" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["subject"] = re.findall(pattern, content)[0].split("<br>")
    
    pattern = r'<td class="Bb font_bold" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["subject_criteria"] = re.findall(pattern, content)[0].split("<br>")
    
    pattern = r'<td class="RBb font_bold" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["subject_magnification"] = re.findall(pattern, content)[0].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["subject_scoring"] = re.findall(pattern, content)[0].split("<br>")

    try:
        pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
        data["subject_score_proportion"] = re.findall(r'>(.*?)<', re.findall(pattern, content)[1])[0]
    except:
        pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
        data["subject_score_proportion"] = re.findall(pattern, content)[1]

    pattern = r'<td class="Bb" rowspan="7" style="text-align:left;vertical-align:text-top;">(.*?)</td>'
    data["specify_items"] = re.findall(pattern, content)[0].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["specify_items_criteria"] = re.findall(pattern, content)[2].split("<br>")

    pattern = r'<td class="RBb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["specify_items_score_proportion"] = re.findall(pattern, content)[0].split("<br>")

    pattern = r'<td class="Rb" rowspan="4" style="text-align:left;vertical-align:text-top;">(.*?)</td>'
    data["excessive_screening"] = re.findall(pattern, content)[0].split("<br>")

    return data

def APCS_group(content):
    data = {"tag": "APCS"}

    pattern = r'<td class="Rb" style="color:#FF0000;width:40px;">(.*?)</td>'
    data["code"] = re.findall(pattern, content)[0]

    pattern = r'<div class="colname">(.*?)</div>'
    data["school_name"] = re.findall(pattern, content)[0]
    
    pattern = r'<div class="gsdname">(.*?)</div>'
    data["department_name"] = re.findall(pattern, content)[0]
    
    pattern = r'<td class="Rb">(.*?)</td>'
    data["enrollment_quota"] = re.findall(pattern, content)[0]
    
    pattern = r'<td class="Rb" style="line-height:10px;">(.*?)</td>'
    data["screening_date"] = format_dates(re.findall(pattern, content)[0])
    
    pattern = r'<td class="Bb font_bold" colspan="2" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["subject"] = re.findall(pattern, content)[0].split("<br>")
    
    pattern = r'<td class="Bb font_bold" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["subject_criteria"] = re.findall(pattern, content)[0].split("<br>")
    
    pattern = r'<td class="RBb font_bold" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["subject_magnification"] = re.findall(pattern, content)[0].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["subject_scoring"] = re.findall(pattern, content)[0].split("<br>")

    try:
        pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
        data["subject_score_proportion"] = re.findall(r'>(.*?)<', re.findall(pattern, content)[1])[0]
    except:
        pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
        data["subject_score_proportion"] = re.findall(pattern, content)[1]

    pattern = r'<td class="Bb" rowspan="7" style="text-align:left;vertical-align:text-top;">(.*?)</td>'
    data["specify_items"] = re.findall(pattern, content)[0].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["specify_items_criteria"] = re.findall(pattern, content)[2].split("<br>")

    pattern = r'<td class="RBb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["specify_items_score_proportion"] = re.findall(pattern, content)[0].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["APCS_items"] = re.findall(pattern, content)[3].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["APCS_items_criteria"] = re.findall(pattern, content)[4].split("<br>")

    pattern = r'<td class="RBb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["APCS_magnification"] = re.findall(pattern, content)[1].split("<br>")
    
    pattern = r'<td class="Rb" rowspan="4" style="text-align:left;vertical-align:text-top;">(.*?)</td>'
    data["excessive_screening"] = re.findall(pattern, content)[0].split("<br>")

    return data

def professional_skills_group(content):
    data = {"tag": "professional_skills"}

    pattern = r'<td class="Rb" style="color:#FF0000;width:59px;">(.*?)</td>'
    data["code"] = re.findall(pattern, content)[0]

    pattern = r'<div class="colname">(.*?)</div>'
    data["school_name"] = re.findall(pattern, content)[0]
    
    pattern = r'<div class="gsdname">(.*?)</div>'
    data["department_name"] = re.findall(pattern, content)[0]
    
    pattern = r'<td class="Rb">(.*?)</td>'
    data["enrollment_quota"] = re.findall(pattern, content)[0]
    
    pattern = r'<td class="Rb" style="line-height:10px;">(.*?)</td>'
    data["screening_date"] = format_dates(re.findall(pattern, content)[0])
    
    pattern = r'<td class="Bb font_bold" colspan="2" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["subject"] = re.findall(pattern, content)[0].split("<br>")
    if data["subject"] == ['--']:
        data["subject_criteria"] = ['--']
        data["subject_magnification"] = ['--']
        data["subject_scoring"] = ['--']
        data["subject_score_proportion"] = ['--']

    else:
        pattern = r'<td class="Bb font_bold" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
        data["subject_criteria"] = re.findall(pattern, content)[0].split("<br>")
        
        pattern = r'<td class="RBb font_bold" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
        data["subject_magnification"] = re.findall(pattern, content)[0].split("<br>")

        pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
        data["subject_scoring"] = re.findall(pattern, content)[0].split("<br>")

        try:
            pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
            data["subject_score_proportion"] = re.findall(r'>(.*?)<', re.findall(pattern, content)[1])[0]
        except:
            pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
            data["subject_score_proportion"] = re.findall(pattern, content)[1]
            
    pattern = r'<td class="Bb" rowspan="7" style="text-align:left;vertical-align:text-top;">(.*?)</td>'
    data["specify_items"] = re.findall(pattern, content)[0].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["specify_items_criteria"] = re.findall(pattern, content)[2].split("<br>")

    pattern = r'<td class="RBb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["specify_items_score_proportion"] = re.findall(pattern, content)[0].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["professional_skills_items"] = re.findall(pattern, content)[3].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["professional_skills_items_criteria"] = re.findall(pattern, content)[4].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["professional_skills_magnification"] = re.findall(pattern, content)[5].split("<br>")

    pattern = r'<td class="Bb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    data["professional_skills_scoring"] = re.findall(pattern, content)[6].split("<br>")

    pattern = r'<td class="RBb" rowspan="7" style="vertical-align:text-top;">(.*?)</td>'
    matches = re.findall(r'>(.*?)<', re.findall(pattern, content)[1])
    data["professional_skills_score_proportion"] = matches[0] if matches else re.findall(pattern, content)[1]


    
    pattern = r'<td class="Rb" rowspan="4" style="text-align:left;vertical-align:text-top;">(.*?)</td>'
    data["excessive_screening"] = re.findall(pattern, content)[0].split("<br>")

    return data

def main():
    if not os.path.exists(folder_name):
        print(f"Folder {folder_name} does not exist!")
    else:
        os.makedirs(output_folder, exist_ok=True)
        os.makedirs(log_folder, exist_ok=True)

        operations = [general_group, APCS_group, professional_skills_group]

        for filename in os.listdir(folder_name):
            file_path = os.path.join(folder_name, filename)
            
            if os.path.isfile(file_path):
                with open(file_path, 'r', encoding='utf-8') as file:
                    content = file.read()
                    success = False
                    error_message = {}

                    for op in operations:
                        try:
                            data = op(content)
                            create_json_file(data)
                            success = True

                        except Exception as e:
                            error_message[op] = e
                    
                    if not success:
                        error_message_string = f"{datetime.now()} - Runtime error! - {file_path}\n{error_message}"
                        log_file_path = os.path.join(log_folder, "info_retriever.log")
                        with open(log_file_path, 'a', encoding='utf-8') as log_file:
                            log_file.write(error_message_string + '\n\n')
                        print(f"Error logged in: {log_file_path}")

if __name__ == "__main__":
    #main()
    
    with open("./file/original/112/001012.html", 'r', encoding='utf-8') as file:
        content = file.read()
        general_group(content)
        #APCS_group(content)
        #professional_skills_group(content)
    