<!DOCTYPE html>
<html>
	<head>
		<title>cs-rsync云存储同步工具</title>
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<!-- Bootstrap -->
		<link rel="stylesheet" href="/public/css/bootstrap.min.css">

		<!-- jQuery (necessary for Bootstrap's JavaScript plugins) -->
		<script src="http://cdn.bootcss.com/jquery/1.10.2/jquery.min.js"></script>
		<!-- Include all compiled plugins (below), or include individual files as needed -->
		<script src="/public/js/bootstrap.min.js"></script>

		<!-- HTML5 Shim and Respond.js IE8 support of HTML5 elements and media queries -->
		<!-- WARNING: Respond.js doesn't work if you view the page via file:// -->
		<!--[if lt IE 9]>
			<script src="http://cdn.bootcss.com/html5shiv/3.7.0/html5shiv.min.js"></script>
			<script src="http://cdn.bootcss.com/respond.js/1.3.0/respond.min.js"></script>
		<![endif]-->
	</head>
	<body>
		<div class="container-fluid">
			<div class="row-fluid">
				<div class="span12">
					<div class="navbar">
						<div class="navbar-inner">
							<div class="container-fluid">
								 <a data-target=".navbar-responsive-collapse" data-toggle="collapse" class="btn btn-navbar"><span class="icon-bar"></span><span class="icon-bar"></span><span class="icon-bar"></span></a> <a href="/" class="brand">cs-rsync云存储同步工具</a>
								<div class="nav-collapse collapse navbar-responsive-collapse in">
									<ul class="nav">
										<li><a href="/SyncInfo">同步信息</a></li>
										<li><a href="/SyncStart">开始同步</a></li>
										<li><a href="/SyncStop">暂停同步</a></li>
										<li><a href="/RemoteBucket">Bucket信息</a></li>
									</ul>
									<ul class="nav pull-right" >
										<li class="divider-vertical"></li>
										<li class="dropdown"> <a data-toggle="dropdown" class="dropdown-toggle" href="#">{{.Username}} <b class="caret"></b></a>
											<ul class="dropdown-menu">
											<li><a href="/Logout">退出登录</a></li>
											<li><a href="/UpdatePassword">修改密码</a></li>
											<li class="divider"></li>
											<li><a href="/Quit">退出程序</a></li>
											</ul>
										</li>
			                        </ul>
								</div>
							</div>
						</div>
					</div>	
				</div>
			</div>
		</div>
