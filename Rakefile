require './util/rake'

GOPATH=[
	'build/go',
	'go'
]

ENV['GOPATH'] = GOPATH.map { |p|
	"#{Dir.pwd}/#{p}"
}.join(':')

godeps = [
	'github.com/golang/protobuf/...',
	'github.com/jordanorelli/moon'
].map { |pkg|
	path = pkg.gsub(/\/\.\.\.$/, '')
	dest = "build/go/src/#{path}"
	file dest do
		sh 'go', 'get', pkg
	end
	dest
}

task :default => [
	'build/bin/server',
	'build/bin/client',
]

task :test do
	sh 'go', 'test', 'pypibot/...'
end

task :edit => godeps do
	sh 'subl', '.'
end

file 'build/bin/server' => godeps + FileList['go/src/pypibot/**/*'] do |t|
	sh 'go', 'build', '-o', t.name, 'pypibot/server'
end

file 'build/bin/client' => godeps + FileList['go/src/pypibot/**/*'] do |t|
	sh 'go', 'build', '-o', t.name, 'pypibot/client'
end

task :nuke do
	rm_rf 'build'
end