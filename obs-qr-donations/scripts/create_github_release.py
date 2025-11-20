#!/usr/bin/env python3
"""
GitHub Release Creator for OBS QR Donations Plugin

This script automates the process of creating a GitHub release with all the
necessary assets (installers, documentation, etc.)
"""
import os
import sys
import json
import requests
from pathlib import Path
from datetime import datetime

# Configuration
REPO_OWNER = "your-username"
REPO_NAME = "obs-qr-donations"
VERSION = "1.0.0"  # This would typically come from a config file or git tag
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN")

class GitHubRelease:
    def __init__(self, token, owner, repo):
        self.token = token
        self.owner = owner
        self.repo = repo
        self.headers = {
            "Authorization": f"token {token}",
            "Accept": "application/vnd.github.v3+json"
        }
        self.api_url = f"https://api.github.com/repos/{owner}/{repo}"
    
    def create_release(self, tag_name, name, body, draft=False, prerelease=False):
        """Create a new GitHub release."""
        url = f"{self.api_url}/releases"
        
        data = {
            "tag_name": tag_name,
            "name": name,
            "body": body,
            "draft": draft,
            "prerelease": prerelease
        }
        
        response = requests.post(url, headers=self.headers, json=data)
        response.raise_for_status()
        return response.json()
    
    def upload_asset(self, release_id, file_path):
        """Upload an asset to a release."""
        url = f"{self.api_url}/releases/{release_id}/assets"
        params = {"name": os.path.basename(file_path)}
        
        with open(file_path, 'rb') as f:
            response = requests.post(
                f"{url}?name={params['name']}",
                headers={
                    **self.headers,
                    "Content-Type": "application/octet-stream"
                },
                data=f.read()
            )
        
        response.raise_for_status()
        return response.json()

def generate_release_notes():
    """Generate release notes from CHANGELOG.md."""
    changelog_path = Path("CHANGELOG.md")
    if not changelog_path.exists():
        return f"Release {VERSION}"
    
    with open(changelog_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()
    
    # Find the current version section
    version_header = f"## [{VERSION}]"
    in_version = False
    notes = []
    
    for line in lines:
        if line.startswith(version_header):
            in_version = True
            continue
        if in_version and line.startswith('## ['):  # Next version
            break
        if in_version:
            notes.append(line)
    
    return "".join(notes).strip() or f"Release {VERSION}"

def get_installer_files():
    """Get a list of installer files to upload."""
    dist_dir = Path("dist")
    if not dist_dir.exists():
        return []
    
    installers = []
    for ext in ['.exe', '.dmg', '.deb', '.tar.gz', '.zip']:
        installers.extend(list(dist_dir.glob(f"*{ext}")))
    
    return installers

def main():
    if not GITHUB_TOKEN:
        print("Error: GITHUB_TOKEN environment variable not set")
        sys.exit(1)
    
    # Get release notes
    release_notes = generate_release_notes()
    tag_name = f"v{VERSION}"
    release_name = f"OBS QR Donations {VERSION}"
    
    # Get installer files
    installers = get_installer_files()
    if not installers:
        print("No installer files found in dist/ directory")
        sys.exit(1)
    
    print(f"Creating release {tag_name} with {len(installers)} assets...")
    
    # Create GitHub release
    github = GitHubRelease(GITHUB_TOKEN, REPO_OWNER, REPO_NAME)
    
    try:
        # Create the release
        release = github.create_release(
            tag_name=tag_name,
            name=release_name,
            body=release_notes,
            draft=True  # Create as draft first, then publish after uploading assets
        )
        
        print(f"Created release: {release['html_url']}")
        
        # Upload assets
        for installer in installers:
            print(f"Uploading {installer.name}...")
            try:
                asset = github.upload_asset(release["id"], str(installer))
                print(f"  ✓ {asset['name']} ({asset['size'] / 1024 / 1024:.2f} MB)")
            except Exception as e:
                print(f"  ✗ Failed to upload {installer.name}: {str(e)}")
        
        # Update release to published status
        update_url = f"{github.api_url}/releases/{release['id']}"
        response = requests.patch(
            update_url,
            headers=github.headers,
            json={"draft": False}
        )
        response.raise_for_status()
        
        print("\n🎉 Release published successfully!")
        print(f"🔗 {release['html_url']}")
        
    except Exception as e:
        print(f"Error creating release: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    main()
