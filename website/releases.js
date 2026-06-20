(function () {
  var REPO = 'merl111/nanvil';
  var RELEASES_URL = 'https://github.com/' + REPO + '/releases';
  var API_URL = 'https://api.github.com/repos/' + REPO + '/releases/latest';

  var versionEl = document.getElementById('release-version');
  var primaryEl = document.getElementById('release-primary');
  var allEl = document.getElementById('release-all');
  var installEl = document.getElementById('release-install-cmd');
  var fallbackEl = document.getElementById('release-fallback');
  var panelEl = document.getElementById('release-panel');

  if (!versionEl || !primaryEl) {
    return;
  }

  function detectPlatform() {
    var ua = navigator.userAgent || '';
    var platform = (navigator.userAgentData && navigator.userAgentData.platform) || navigator.platform || '';
    var lower = (ua + ' ' + platform).toLowerCase();

    if (lower.indexOf('win') !== -1) {
      return { goos: 'windows', goarch: 'amd64', label: 'Windows (amd64)', ext: 'zip' };
    }
    if (lower.indexOf('mac') !== -1 || lower.indexOf('darwin') !== -1) {
      var arm = lower.indexOf('arm') !== -1 || (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1);
      if (arm) {
        return { goos: 'darwin', goarch: 'arm64', label: 'macOS (Apple Silicon)', ext: 'tar.gz' };
      }
      return { goos: 'darwin', goarch: 'amd64', label: 'macOS (Intel)', ext: 'tar.gz' };
    }
    if (lower.indexOf('aarch64') !== -1 || lower.indexOf('arm64') !== -1) {
      return { goos: 'linux', goarch: 'arm64', label: 'Linux (arm64)', ext: 'tar.gz' };
    }
    return { goos: 'linux', goarch: 'amd64', label: 'Linux (amd64)', ext: 'tar.gz' };
  }

  function findAsset(assets, goos, goarch) {
    var suffixTar = '-' + goos + '-' + goarch + '.tar.gz';
    var suffixZip = '-' + goos + '-' + goarch + '.zip';
    for (var i = 0; i < assets.length; i++) {
      var name = assets[i].name || '';
      if (name.indexOf(suffixTar) !== -1 || name.indexOf(suffixZip) !== -1) {
        return assets[i];
      }
    }
    return null;
  }

  function installSnippet(asset, plat) {
    var file = asset.name;
    if (plat.goos === 'windows') {
      return '# Download and extract ' + file + '\n# Then run: .\\nanvil.exe start';
    }
    return 'curl -sSL ' + asset.browser_download_url + ' | tar xz\n./nanvil start';
  }

  function showFallback(message) {
    if (panelEl) {
      panelEl.classList.add('release-panel--fallback');
    }
    versionEl.textContent = message || 'No release yet';
    primaryEl.textContent = 'View releases on GitHub';
    primaryEl.href = RELEASES_URL;
    if (allEl) {
      allEl.href = RELEASES_URL;
    }
    if (installEl) {
      installEl.textContent = 'git clone https://github.com/' + REPO + '.git\n'
        + 'cd nanvil && make build\n./bin/nanvil start';
    }
    if (fallbackEl) {
      fallbackEl.hidden = false;
    }
  }

  function showRelease(data) {
    var plat = detectPlatform();
    var asset = findAsset(data.assets || [], plat.goos, plat.goarch);
    var tag = data.tag_name || 'latest';

    versionEl.textContent = tag;
    if (allEl) {
      allEl.href = data.html_url || RELEASES_URL;
    }

    if (!asset) {
      primaryEl.textContent = 'Download ' + tag;
      primaryEl.href = data.html_url || RELEASES_URL;
      if (installEl) {
        installEl.textContent = '# Pick your platform from the release page:\n' + (data.html_url || RELEASES_URL);
      }
      return;
    }

    primaryEl.textContent = 'Download for ' + plat.label;
    primaryEl.href = asset.browser_download_url;
    if (installEl) {
      installEl.textContent = installSnippet(asset, plat);
    }
    if (fallbackEl) {
      fallbackEl.hidden = false;
    }
  }

  fetch(API_URL, { headers: { Accept: 'application/vnd.github+json' } })
    .then(function (resp) {
      if (resp.status === 404) {
        showFallback('No release published yet');
        return null;
      }
      if (!resp.ok) {
        throw new Error('GitHub API ' + resp.status);
      }
      return resp.json();
    })
    .then(function (data) {
      if (data) {
        showRelease(data);
      }
    })
    .catch(function () {
      showFallback('Could not load latest release');
    });
})();
