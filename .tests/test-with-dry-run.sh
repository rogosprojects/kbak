#!/bin/bash

# Treat unset variables as an error
set -u

# Error handling
handle_error() {
  local exit_code=$?
  print_error "Command failed with exit code $exit_code"
  # Continue execution instead of exiting
}

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
check_command() {
  if ! command -v "$1" &> /dev/null; then
    print_error "$1 command not found. Please install it."
    exit 1
  fi
}

# Configuration
TEST_NAMESPACE=${1:-"test"}
OUTPUT_DIR="kbak_test_output"
BUILD_DIR=".."

main() {
  print_info "Starting kbak test for namespace: ${TEST_NAMESPACE}"

  # Check for required commands
  check_command "kubectl"

  # Step 1: Build kbak if binary doesn't exist
  if [ ! -f "${BUILD_DIR}/kbak" ]; then
    print_info "Building kbak..."
    check_command "go"
    cd "${BUILD_DIR}" && go build -o kbak ./cmd/kbak
    if [ $? -ne 0 ]; then
      print_error "Failed to build kbak."
      exit 1
    fi
    print_success "kbak built successfully."
  else
    print_info "Using existing kbak binary."
  fi

  # Step 2: Run kbak to backup the namespace
  print_info "Running kbak to backup namespace ${TEST_NAMESPACE}..."
  "${BUILD_DIR}/kbak" --namespace "${TEST_NAMESPACE}" --output "${OUTPUT_DIR}" --verbose
  if [ $? -ne 0 ]; then
    print_error "Failed to run kbak backup."
    exit 1
  fi

  # Get the latest backup directory by timestamp
  LATEST_BACKUP=$(find "${OUTPUT_DIR}" -mindepth 1 -maxdepth 1 -type d | sort | tail -n 1)
  if [ -z "$LATEST_BACKUP" ]; then
    print_error "No backup directory found in ${OUTPUT_DIR}"
    exit 1
  fi

  BACKUP_DIR="${LATEST_BACKUP}/${TEST_NAMESPACE}"
  if [ ! -d "${BACKUP_DIR}" ]; then
    print_error "Backup directory not found: ${BACKUP_DIR}"
    exit 1
  fi

  print_success "Backup created at: ${BACKUP_DIR}"

  # Step 3: Validate backup contents
  print_info "Validating backup contents..."

  # Check if any resources were backed up
  RESOURCE_DIRS=$(find "${BACKUP_DIR}" -mindepth 1 -maxdepth 1 -type d)
  if [ -z "${RESOURCE_DIRS}" ]; then
    print_warning "No resources found in backup directory. Is the namespace empty?"
    exit 0
  fi

  RESOURCE_DIR_COUNT=$(echo "$RESOURCE_DIRS" | wc -l)
  print_info "Resource types found: ${RESOURCE_DIR_COUNT}"

  # Test each resource type directory
  TOTAL_RESOURCES=0
  SUCCESSFUL_RESOURCES=0
  FAILED_RESOURCES=0

  for DIR in $RESOURCE_DIRS; do
    if [ ! -d "$DIR" ]; then
      continue
    fi

    RESOURCE_TYPE=$(basename "${DIR}")
    RESOURCE_FILES=$(find "${DIR}" -mindepth 1 -maxdepth 1 -type f -name "*.yaml")

    # Skip if no yaml files found
    if [ -z "$RESOURCE_FILES" ]; then
      print_warning "No YAML files found in ${RESOURCE_TYPE}"
      continue
    fi

    RESOURCE_COUNT=$(echo "$RESOURCE_FILES" | wc -l)
    print_info "Testing ${RESOURCE_COUNT} ${RESOURCE_TYPE} resources..."

    for FILE in $RESOURCE_FILES; do
      if [ ! -f "$FILE" ]; then
        continue
      fi

      RESOURCE_NAME=$(basename "${FILE}" .yaml)
      ((TOTAL_RESOURCES++))

      # Validate each resource with kubectl dry-run
      if kubectl apply -f "${FILE}" --dry-run=client &> /tmp/kbak_test_result; then
        print_success "Resource validated: ${RESOURCE_TYPE}/${RESOURCE_NAME}"
        ((SUCCESSFUL_RESOURCES++))
      else
        print_error "Resource validation failed: ${RESOURCE_TYPE}/${RESOURCE_NAME}"
        print_error "$(cat /tmp/kbak_test_result)"
        ((FAILED_RESOURCES++))
      fi
    done
  done

  # Step 4: Try to validate entire resource types at once
  print_info "Validating entire resource types at once..."

  for DIR in $RESOURCE_DIRS; do
    if [ ! -d "$DIR" ]; then
      continue
    fi

    RESOURCE_TYPE=$(basename "${DIR}")

    # Check if directory has YAML files
    YAML_FILES=$(find "${DIR}" -mindepth 1 -maxdepth 1 -type f -name "*.yaml")
    if [ -z "$YAML_FILES" ]; then
      print_warning "No YAML files found in ${RESOURCE_TYPE} - skipping bulk validation"
      continue
    fi

    # Validate entire resource type directory
    if kubectl apply -f "${DIR}" --dry-run=client &> /tmp/kbak_test_result; then
      print_success "Resource type validated: ${RESOURCE_TYPE}"
    else
      print_error "Resource type validation failed: ${RESOURCE_TYPE}"
      print_error "$(cat /tmp/kbak_test_result)"
    fi
  done

  # Step 5: Summary
  echo ""
  echo "=============================================="
  echo "              TEST SUMMARY                   "
  echo "=============================================="
  echo "Total resources tested: ${TOTAL_RESOURCES}"
  echo "Successful validations: ${SUCCESSFUL_RESOURCES}"
  echo "Failed validations: ${FAILED_RESOURCES}"
  echo "=============================================="

  if [ ${FAILED_RESOURCES} -eq 0 ]; then
    print_success "All resources validated successfully!"
    exit 0
  else
    print_error "${FAILED_RESOURCES} resources failed validation."
    exit 1
  fi
}

# Run the main function
main