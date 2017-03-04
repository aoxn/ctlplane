podTemplate(label: 'mypod', inheritFrom: 'jnlp', containers: [
    containerTemplate(name: 'golang', image: 'registry.cn-hangzhou.aliyuncs.com/spacexnice/golang:1.6.3', ttyEnabled: true, command: 'cat'),
    containerTemplate(name: 'docker', image: 'registry.cn-hangzhou.aliyuncs.com/spacexnice/docker:1.12.6', command: 'cat', ttyEnabled: true)
  ]
  ,volumes: [
/*
      persistentVolumeClaim(mountPath: '/home/jenkins', claimName: 'jenkins', readOnly: false),
*/
     hostPathVolume(hostPath: '/var/run/docker.sock', mountPath: '/var/run/docker.sock')
  ]) {

    node ('jenkins-pod-slave') {

        stage('Get a Golang project') {
            git url: 'https://github.com/spacexnice/blog.git'
            container('docker') {
                stage('Build blog project') {
                    sh """
                        docker build -t registry.cn-hangzhou.aliyuncs.com/spacexnice/blog:v1.0.0 .
                        """

                }
            }
        }

        stage 'Build and push images!'
        container('docker') {
            stage 'build'
            /*
            withDockerRegistry([credentialsId: '12e859f9-bba6-4963-ac38-975ca794e58e', url: 'https://registry.cn-hangzhou.aliyuncs.com']) {
                sh """
                        docker push registry.cn-hangzhou.aliyuncs.com/spacexnice/blog:v1.0.0
                """
            }
            */
        }
    }
}
