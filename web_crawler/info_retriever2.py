from bs4 import BeautifulSoup
import csv
import os

# Directory and file settings
input_directory = 'file/original/113_'
output_directory = 'file/csv/113_'

# Ensure the output directory exists
os.makedirs(output_directory, exist_ok=True)

# Function to process a single HTML file
def process_html_file(file_path):
    # Read the HTML file
    with open(file_path, 'r', encoding='utf-8') as file:
        soup = BeautifulSoup(file, 'html.parser')

    # Find all table rows (excluding the header)
    rows = soup.find_all('tr')[1:]

    # Process each row
    for row in rows:
        cells = row.find_all('td')
        if cells:
            department_code = cells[0].text.strip()
            if department_code:  # Skip rows with empty department code
                department_name = cells[1].text.strip()
                quota_type = cells[2].text.strip()
                technical_item = cells[3].text.strip()
                gender_restriction = cells[4].text.strip()
                minimum_standard = cells[5].text.strip()

                # Check the conditions for quota_type, technical_item, and gender_restriction
                if quota_type != '招生' or technical_item != '--' or gender_restriction != '無':
                    department_name = '--'
                    quota_type = '--'
                    technical_item = '--'
                    gender_restriction = '--'
                    minimum_standard = '--'

                # Prepare the output file path
                output_file_path = os.path.join(output_directory, f"{department_code}.csv")

                # Write the data to a CSV file
                with open(output_file_path, 'w', newline='', encoding='utf-8') as csv_file:
                    writer = csv.writer(csv_file)
                    writer.writerow([
                        department_code,
                        department_name,
                        quota_type,
                        technical_item,
                        gender_restriction,
                        minimum_standard
                    ])

# Process all HTML files in the input directory
for filename in os.listdir(input_directory):
    if filename.endswith('.html'):
        file_path = os.path.join(input_directory, filename)
        process_html_file(file_path)
