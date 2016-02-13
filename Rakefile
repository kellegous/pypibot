require './util/rake'

EDITOR='atom'

GOPATH=[
	'build/go',
	'.'
]

ENV['GOPATH'] = GOPATH.map { |p|
	File.absolute_path(File.join(Dir.pwd, p))
}.join(':')

ENV['PATH'] = "#{Dir.pwd}/build/go/bin:#{ENV['PATH']}"

godeps = go_get('build/go/src', [
	'github.com/golang/protobuf/...',
	'github.com/syndtr/goleveldb/leveldb',
	'github.com/scalingdata/gcfg'
])

protobufs = protoc('src')

core = godeps + protobufs

task :default => [
	'build/bin/server',
	'build/bin/client',
]

task :test do
	sh 'go', 'test', 'pypibot/...'
end

task :edit => godeps do
	sh EDITOR, '.'
end

file 'build/bin/server' => core + FileList['go/src/pypibot/**/*'] do |t|
	sh 'go', 'build', '-o', t.name, 'pypibot/server'
end

file 'build/bin/client' => core + FileList['go/src/pypibot/**/*'] do |t|
	sh 'go', 'build', '-o', t.name, 'pypibot/client'
end

task :nuke do
	rm_rf 'build'
end

task :zing do
	protoc2('src', 'build')
end
