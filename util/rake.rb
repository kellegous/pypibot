require 'fileutils'

# Allows auto generation of a clean task from file declarations
CLEAN = Rake::FileList["**/*~", "**/*.bak", "**/core"]

# Invoke a shell command and wait for it to complete successfully
def call(*args)
  Process.wait spawn(*args)
  fail "command status: #{s} for #{args.inspect}" unless $?.exitstatus == 0
end

# A file task that marks the target for removal during clean.
def file!(*args, &block)
  task = Rake::FileTask.define_task(*args, &block)
  CLEAN.include(task.name)
  task
end

# A clean task that removes everything that was added to CLEAN
task :clean do
  CLEAN.each { |f| rm_r(f) rescue nil }
end

# generates build rules for protobufs. Rules that target dst are generated
# by scanning src.
def protoc(src, dst)
	file dst do
		FileUtils.mkdir_p dst
	end

	FileList["#{src}/*.proto"].map do |proto|
		dest = File.join(dst, File.basename(proto, '.proto') + '.pb.go')
		file dest => [proto, dst] do
			sh 'protoc', "-I#{src}", "--go_out=#{dst}", proto
		end

		dest
	end
end

def go_get(dst, deps)
	deps.map do |pkg|
		path = pkg.gsub(/\/\.\.\.$/, '')
		dest = File.join(dst, path)
		file dest do
			sh 'go', 'get', pkg
		end
		dest
	end
end