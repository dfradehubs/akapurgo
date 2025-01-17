document.getElementById('purge-form').addEventListener('submit', async function(event) {
    event.preventDefault();
    const messageElement = document.getElementById('message');
    messageElement.textContent = '';

    // Get form data
    const purgeType = document.getElementById('purge-type').value;
    const actionType = document.getElementById('action-type').value;
    const environment = document.getElementById('environment').value;
    const paths = document.getElementById('paths').value.trim().split('\n').filter(Boolean);

    if (paths.length === 0) {
        messageElement.textContent = 'Please enter at least one path or tag.';
        messageElement.className = 'message error';
        return;
    }

    try {
        const response = await fetch('/api/v1/purge', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                purgeType,
                actionType,
                environment,
                paths
            })
        });

        if (response.ok) {
            messageElement.textContent = 'Cache purged successfully.';
            messageElement.className = 'message success';
        } else {
            const errorData = await response.json();
            messageElement.textContent = `Error: ${errorData.message || 'Failed to purge cache.'}`;
            messageElement.className = 'message error';
        }
    } catch (error) {
        messageElement.textContent = 'An unexpected error occurred. Please try again.';
        messageElement.className = 'message error';
    }
});
