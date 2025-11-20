#!/usr/bin/env python3
"""
OBS QR Donations Plugin Installation Test

This script verifies the installation of the OBS QR Donations plugin.
It checks for required files, dependencies, and basic functionality.
"""
import os
import sys
import platform
import json
from pathlib import Path

class InstallationTester:
    def __init__(self):
        self.platform = platform.system()
        self.obs_plugin_dir = self.get_obs_plugin_dir()
        self.plugin_dir = self.obs_plugin_dir / "qr-donations"
        self.errors = []
        self.warnings = []
    
    def get_obs_plugin_dir(self):
        """Get the OBS plugin directory for the current platform."""
        if self.platform == "Windows":
            return Path(os.path.expandvars(r"%APPDATA%\\obs-studio\\plugins"))
        elif self.platform == "Darwin":
            return Path("~/Library/Application Support/obs-studio/plugins").expanduser()
        else:  # Linux
            return Path("~/.config/obs-studio/plugins").expanduser()
    
    def check_plugin_directory(self):
        """Check if the plugin directory exists and has the correct structure."""
        print("\n🔍 Checking plugin directory...")
        
        if not self.plugin_dir.exists():
            self.errors.append(f"Plugin directory not found: {self.plugin_dir}")
            return False
        
        print(f"✅ Found plugin directory: {self.plugin_dir}")
        return True
    
    def check_required_files(self):
        """Check for required plugin files."""
        print("\n📁 Checking required files...")
        
        required_files = [
            "bin/obs-qr-donations.dll" if self.platform == "Windows" else "bin/obs-qr-donations.so",
            "data/settings.json",
            "README.txt"
        ]
        
        all_files_exist = True
        for file in required_files:
            file_path = self.plugin_dir / file
            if not file_path.exists():
                self.errors.append(f"Missing required file: {file}")
                all_files_exist = False
            else:
                print(f"✅ Found: {file}")
        
        return all_files_exist
    
    def check_obs_version(self):
        """Check if OBS Studio version is compatible."""
        print("\n📊 Checking OBS Studio version...")
        
        try:
            import obspython as obs
            version = obs.obs_get_version_string()
            print(f"✅ OBS Studio version: {version}")
            
            # Check if version is 28.0 or later
            major_version = int(version.split('.')[0])
            if major_version < 28:
                self.warnings.append(f"OBS Studio version {version} may not be compatible. Version 28.0 or later is recommended.")
                return False
            return True
            
        except ImportError:
            self.errors.append("Could not import obspython. Make sure OBS Studio is installed and this script is running in the OBS Python environment.")
            return False
    
    def run_tests(self):
        """Run all installation tests."""
        print("🚀 Starting OBS QR Donations Plugin Installation Test\n")
        
        tests = [
            ("OBS Studio Version", self.check_obs_version),
            ("Plugin Directory", self.check_plugin_directory),
            ("Required Files", self.check_required_files)
        ]
        
        results = {}
        for name, test_func in tests:
            print(f"\n=== {name} ===")
            try:
                result = test_func()
                results[name] = "PASSED" if result else "FAILED"
            except Exception as e:
                results[name] = f"ERROR: {str(e)}"
                self.errors.append(f"Error in {name}: {str(e)}")
        
        # Print summary
        print("\n📋 Test Summary:")
        print("-" * 40)
        for name, result in results.items():
            status = "✅ PASSED" if result == "PASSED" else f"❌ {result}"
            print(f"{name}: {status}")
        
        # Print warnings and errors
        if self.warnings:
            print("\n⚠️  Warnings:")
            for warning in self.warnings:
                print(f"  • {warning}")
        
        if self.errors:
            print("\n❌ Errors:")
            for error in self.errors:
                print(f"  • {error}")
            print("\n❌ Installation test failed. Please check the errors above.")
            return False
        
        print("\n🎉 All tests passed! The plugin is properly installed.")
        return True

def main():
    tester = InstallationTester()
    success = tester.run_tests()
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()
