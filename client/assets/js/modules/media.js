// 媒体模块
export class MediaModule {
    constructor(apiClient, store) {
        this.apiClient = apiClient;
        this.store = store;
    }

    // 上传文件
    async uploadFile(file, type) {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('type', type);

        const response = await this.apiClient.post('/api/media/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        });

        // 缓存媒体信息
        await this.store.media.add(response);

        return response;
    }

    // 上传图片
    async uploadImage(file) {
        if (!file.type.startsWith('image/')) {
            throw new Error('不是图片文件');
        }

        // 验证大小 (5MB)
        if (file.size > 5 * 1024 * 1024) {
            throw new Error('图片太大 (最大 5MB)');
        }

        return this.uploadFile(file, 'image');
    }

    // 上传语音
    async uploadVoice(file) {
        if (!file.type.startsWith('audio/')) {
            throw new Error('不是音频文件');
        }

        // 验证大小 (10MB)
        if (file.size > 10 * 1024 * 1024) {
            throw new Error('音频太大 (最大 10MB)');
        }

        return this.uploadFile(file, 'voice');
    }

    // 获取媒体URL
    getMediaUrl(mediaId, thumbnail = false) {
        return thumbnail
            ? `/api/media/${mediaId}/thumb`
            : `/api/media/${mediaId}`;
    }

    // 从本地缓存获取媒体信息
    async getCachedMedia(mediaId) {
        return await this.store.media.get(mediaId);
    }

    // 删除媒体
    async deleteMedia(mediaId) {
        await this.apiClient.delete(`/api/media/${mediaId}`);
    }

    // 选择图片文件
    selectImage() {
        return new Promise((resolve, reject) => {
            const input = document.createElement('input');
            input.type = 'file';
            input.accept = 'image/*';
            input.onchange = (e) => {
                const file = e.target.files[0];
                if (file) {
                    resolve(file);
                } else {
                    reject(new Error('未选择文件'));
                }
            };
            input.oncancel = () => reject(new Error('取消选择'));
            input.click();
        });
    }

    // 录制语音
    recordVoice(maxDuration = 60000) {
        return new Promise((resolve, reject) => {
            if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
                reject(new Error('浏览器不支持录音'));
                return;
            }

            navigator.mediaDevices.getUserMedia({ audio: true })
                .then(stream => {
                    const mediaRecorder = new MediaRecorder(stream);
                    const audioChunks = [];

                    mediaRecorder.ondataavailable = (event) => {
                        audioChunks.push(event.data);
                    };

                    // 创建一个Promise来等待录音完成
                    let blobResolver = null;
                    const blobPromise = new Promise((resolve) => {
                        blobResolver = resolve;
                    });

                    mediaRecorder.onstop = () => {
                        const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
                        stream.getTracks().forEach(track => track.stop());
                        blobResolver(audioBlob);
                    };

                    mediaRecorder.onerror = (event) => {
                        stream.getTracks().forEach(track => track.stop());
                        reject(event.error);
                    };

                    mediaRecorder.start();

                    // 设置最大录制时长
                    setTimeout(() => {
                        if (mediaRecorder.state === 'recording') {
                            mediaRecorder.stop();
                        }
                    }, maxDuration);

                    // 返回控制对象，stop()返回Promise
                    resolve({
                        stop: () => {
                            mediaRecorder.stop();
                            return blobPromise;
                        },
                        cancel: () => {
                            if (mediaRecorder.state === 'recording') {
                                mediaRecorder.stop();
                                stream.getTracks().forEach(track => track.stop());
                            }
                            reject(new Error('取消录音'));
                            return blobPromise.catch(() => {});
                        }
                    });
                })
                .catch(reject);
        });
    }

    // 播放音频
    playAudio(url) {
        return new Promise((resolve, reject) => {
            const audio = new Audio(url);

            audio.onended = () => resolve();
            audio.onerror = () => reject(new Error('音频播放失败'));

            audio.play().catch(reject);
        });
    }
}
