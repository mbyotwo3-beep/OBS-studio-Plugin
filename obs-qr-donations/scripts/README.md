# OBS QR Donations - Development Scripts

This directory contains various utility scripts to assist with the development, testing, and deployment of the OBS QR Donations plugin.

## Available Scripts

### Development

- `generate_screenshots.py` - A GUI tool to capture and save screenshots of the plugin UI for documentation.
- `generate_tutorial.py` - Script to help create video tutorials by generating a structured tutorial script.
- `test_installation.py` - Verifies that the plugin is correctly installed and all required files are present.
- `test_local_installation.ps1` - PowerShell script to test the local installation on Windows.

### Build & Deployment

- `create_installers.py` - Creates platform-specific installers for Windows, macOS, and Linux.
- `create_github_release.py` - Automates the process of creating a GitHub release with all necessary assets.

### Testing

- `run_tests.py` - Runs the test suite for the plugin.
- `test_breez_integration.py` - Tests the Breez SDK integration.

## Usage

1. Install the required dependencies:
   ```bash
   pip install -r ../requirements-dev.txt
   ```

2. Run the desired script:
   ```bash
   python scripts/script_name.py
   ```

## Environment Variables

Some scripts may require environment variables to be set. Check the script's documentation for details.

## Contributing

When adding new scripts:

1. Keep them focused on a single task
2. Add proper error handling
3. Include command-line help
4. Document any dependencies
5. Follow the project's coding style

## License

All scripts are licensed under the same [MIT License](../LICENSE) as the main project.
