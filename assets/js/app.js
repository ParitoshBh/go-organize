var uppy = Uppy.Core({
    debug: false
});

uppy.use(Uppy.Dashboard, {
    target: '.uppy-dashboard-container',
    inline: true,
    height: 300,
    hideUploadButton: true,
    hideRetryButton: true,
    proudlyDisplayPoweredByUppy: false
});

uppy.use(Uppy.XHRUpload, {
    endpoint: '/object/create',
    formData: true,
    limit: 1,
    // disable timeout till the time backend can update js of upload status
    timeout: 0,
    getResponseError: function (responseText, response) {
        return new Error(JSON.parse(responseText).message);
    },
    validateStatus: function (statusCode, responseText, response) {
        if (statusCode != 200) {
            return false;
        }

        if (JSON.parse(responseText).status) {
            return true;
        }

        return false;
    }
});

uppy.on('file-added', (file) => {
    uppy.setFileMeta(file.id, {
        bucketPath: document.getElementById('bucketPath').value
    })
});

function submitModal(e) {
    var submitButton = e.querySelector('button[type="submit"]');
    var selectedTab = null;

    var tabs = document.querySelectorAll('#modal-team [data-bs-toggle="tab"]');
    for (let i = 0; i < tabs.length; i++) {
        if (tabs[i].classList.contains('active')) {
            selectedTab = tabs[i];
            break;
        }
    }

    switch (selectedTab.dataset.type) {
        case 'file':
            submitButton.classList.add('btn-loading');

            var promise = null;
            var currentState = uppy.getState();

            if ((currentState.error === undefined) || (currentState.error === null)) {
                promise = uppy.upload();
            } else {
                promise = uppy.retryAll();
            }

            promise.then((result) => {
                if (result.failed.length > 0) {
                    submitButton.classList.remove('btn-loading');
                    uppy.setState({
                        info: {
                            isHidden: true
                        }
                    });
                } else {
                    window.location.reload();
                }
            });
            break;
        case 'folder':
            return true;
    }

    return false;
}