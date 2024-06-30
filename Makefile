.PHONY: test


LOGIN_URL=http://localhost:8080/login
GATEWAY_URL=http://localhost:8080/generate-playlist

auth: login extract_url request_auth_url clean

login:
	@echo "Making request to $(LOGIN_URL)..."
	@curl -s $(LOGIN_URL) -o login_response.json
	@echo "Login response saved to login_response.json"

extract_url: login
	@echo "Extracting URL from login response..."
	@jq -r '.url |= gsub("\\\\u0026"; "&") | .url' login_response.json > extracted_url.txt
	@echo "Extracted URL saved to extracted_url.txt"

request_auth_url: extract_url
	@echo
	@echo "Access the below URL before proceed..."
	@echo
	@cat extracted_url.txt
	@echo
	@echo "Press any key to continue..."
	@read -n 1 -s

send_request:
	@echo "Making request to $(GATEWAY_URL)..."
	@echo
	@curl -X POST $(GATEWAY_URL) \
	     -H "Content-Type: application/json" \
	     -d @request.json
	@echo
	@echo "Playlist created on Spotify."

clean:
	@echo "Cleaning up files..."
	@rm -f login_response.json extracted_url.txt auth_response.json
	@echo "Cleanup completed"

start:
	@echo "Starting the services..."
	@docker-compose up -d --build

stop:
	@echo "Stopping the services..."
	@docker-compose down