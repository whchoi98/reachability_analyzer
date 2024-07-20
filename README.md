# reachability_analyzer
reachability_analyzer는 AWS reachability analyzer를 기반으로 AWS SDK, golang으로 구성되었습니다.
아래는 터미널에서 실행해야 하는 명령어 목록입니다. 이 명령어들은 Go 코드를 실행하기 위해 필요한 패키지를 설치하고 환경을 설정하는 데 필요합니다.

### 1. AWS CLI 설치 및 구성

### AWS CLI 설치

### macOS

```
brew install awscli

```

### Linux

```
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
source ~/.bashrc
aws --version
which aws_completer
export PATH=/usr/local/bin:$PATH
source ~/.bash_profile
complete -C '/usr/local/bin/aws_completer' aws
source ~/.bashrc
aws --version

```

### Windows

AWS CLI MSI 설치 프로그램을 다운로드하고 실행:
[AWS CLI 설치 프로그램](https://awscli.amazonaws.com/AWSCLIV2.msi)

### AWS CLI 구성

AWS CLI가 설치되면 다음 명령을 사용하여 구성합니다:

```
aws configure
```

프롬프트에 따라 AWS Access Key ID, AWS Secret Access Key, Default region name, Default output format을 입력합니다.

### 2. Go 설치

Go가 설치되어 있어야 합니다. 설치되어 있지 않다면 [Go 공식 사이트](https://golang.org/dl/)에서 다운로드하여 설치합니다.

설치가 완료되면 터미널에서 Go 버전을 확인하여 설치가 제대로 되었는지 확인하십시오:

```
go version
```

### 3. Go 모듈 초기화 및 필요한 패키지 설치

프로젝트 디렉터리로 이동하여 Go 모듈을 초기화하고 필요한 패키지를 설치합니다:

```
# git clone
git clone https://github.com/whchoi98/reachability_analyzer.git

# 프로젝트 디렉터리로 이동
cd /path/to/your/project

```

### 4. 코드 실행

터미널에서 코드 파일이 있는 디렉터리로 이동하여 다음 명령을 실행합니다:

```
go run reachability_analyzer.go <region> <source> <destination> <protocol> <port>

```

각 매개변수는 다음과 같이 대체합니다:

- `<region>`: AWS 리전 (예: `ap-northeast-2`)
- `<source>`: 소스 IP 주소 (예: `10.0.0.1`) / 소스 IP 주소는 도메인 입력이 불가능합니다.
- `<destination>`: 목적지 IP 주소 또는 도메인 (예: `example.com`)
- `<protocol>`: 프로토콜 (예: `tcp/udp`)
- `<port>`: 포트 번호 (예: `80`)

예시:

```
go run reachability_analyzer.go ap-northeast-2 10.0.0.1 example.com tcp 80

```

### 5. IAM 권한 설정

이 코드를 실행하는 데 사용되는 AWS 자격 증명은 Network Insights, EC2 인스턴스 및 관련 네트워크 리소스를 설명하고 분석할 수 있는 적절한 IAM 권한을 가지고 있어야 합니다. 다음은 최소 권한을 제공하는 정책 예시입니다:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:CreateNetworkInsightsPath",
                "ec2:StartNetworkInsightsAnalysis",
                "ec2:DescribeNetworkInsightsAnalyses",
                "ec2:DescribeNetworkInterfaces",
                "ec2:DescribeInternetGateways",
                "ec2:DescribeRouteTables"
            ],
            "Resource": "*"
        }
    ]
}

```

이 정책을 IAM 사용자 또는 역할에 추가하여 필요한 권한을 부여할 수 있습니다.

이 단계를 완료하면 Go 코드를 실행하여 AWS Network Insights를 사용해 네트워크 경로를 분석할 수 있습니다.

아래는 실행 결과 예제입니다.

```
$ go run ./reachability_analyzer_table.go ap-northeast-2 10.1.21.101 www.aws.com tcp 80 
Network Insights Path ID: nip-0441341e8619965e1
Analysis in progress...
Analysis in progress...
Path Visualization:
Hop         |Component ID                 |ACL Rule |
Source      |10.1.21.101                  |         |
1           |eni-00076b439caa634df        |N/A      |
2           |sg-03ae5ce58c948e589         |N/A      |
3           |acl-0a43d9a9667bcab81        |allow    |
4           |rtb-0aad6b8b14781f72d        |N/A      |
5           |acl-0a43d9a9667bcab81        |allow    |
6           |eni-05714cdd76f012672        |N/A      |
7           |tgw-attach-0fc0b5e3d583a80cc |N/A      |
8           |tgw-rtb-077278a1b827415c0    |N/A      |
9           |tgw-attach-01d903af9367836f4 |N/A      |
10          |eni-0b685977ab01dc70f        |N/A      |
11          |acl-09e96f538cf333329        |allow    |
12          |rtb-0b65fc177c0a5f0dd        |N/A      |
13          |acl-09e96f538cf333329        |allow    |
14          |eni-07a2405295c45f8c7        |N/A      |
15          |nat-068fa87ad8e67a7cf        |N/A      |
16          |eni-07a2405295c45f8c7        |N/A      |
17          |acl-09e96f538cf333329        |allow    |
18          |rtb-05c93d0c19954a705        |N/A      |
19          |acl-09e96f538cf333329        |allow    |
20          |eni-0b63288dd9e8089e4        |N/A      |
21          |vpce-0c4933f77bab06c7b       |N/A      |
22          |N2S-firewall-ANFW-N2SVPC     |N/A      |
23          |vpce-0c4933f77bab06c7b       |N/A      |
24          |eni-0b63288dd9e8089e4        |N/A      |
25          |acl-09e96f538cf333329        |allow    |
26          |rtb-018d4cedf80e5dfb4        |N/A      |
27          |igw-01ecf45f7eb83c135        |N/A      |
Destination |18.64.8.60                   |         |

Result: SUCCESS
```
